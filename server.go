package gocmdsrv

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

// Callback function signature for a registered command
type CmdCallback func([]string) (*string, error)

// A registered command to execute from the api
type Command struct {

	// Name of the command
	Name string

	// Handler for the command
	Callback CmdCallback
}

// A struct to hold the value of an execute api request
type ExecuteApiRequest struct {

	// command to execute
	Command string `json:"command"`

	// command arguments
	Args []string `json:"args"`
}

// A command server to provide command execution over a rest api
type CmdServer struct {

	// mux router
	Router *mux.Router

	// server bind port
	Port int

	// List of registered commands
	Commands []*Command
}

/*
Create a new command server
*/
func New() *CmdServer {

	// Create server and router
	srv := CmdServer{}
	srv.Router = mux.NewRouter()

	// Mount routes
	srv.Router.HandleFunc("/execute", srv.ApiHandler).Methods("POST")
	srv.Commands = make([]*Command, 0)
	return &srv
}

/*
Start the server with a specified port
*/
func (srv *CmdServer) StartWithPort(port int) error {

	// Convert port int to string
	servePort := fmt.Sprintf(":%d", port)

	// Save and log port
	srv.Port = port
	fmt.Println("Server started on port", srv.Port)

	// Start server listening
	log.Fatal(http.ListenAndServe(servePort, srv.Router))

	return nil
}

/*
Start the server without a specified port
*/
func (srv *CmdServer) Start() error {

	// Create server listener
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	// Save and log port
	srv.Port = listener.Addr().(*net.TCPAddr).Port
	fmt.Println("Server started on port", srv.Port)

	// Start server listening
	log.Fatal(http.Serve(listener, srv.Router))

	return nil
}

/*
Handle inbound http api requests
*/
func (srv *CmdServer) ApiHandler(w http.ResponseWriter, r *http.Request) {

	// read all request body data
	defer r.Body.Close()
	rawdata, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		SendMessage(w, http.StatusBadRequest, "Malformed request")
		return
	}

	// Get the request json data
	req, parseErr := ParseExecuteRequest(rawdata)
	if parseErr != nil {
		SendMessage(w, http.StatusInternalServerError, "Unable to parse api request")
		return
	}

	// Find the command in the server
	var foundCmd *Command
	for _, cmd := range srv.Commands {

		// Find the matching command
		if cmd.Name == req.Command {
			foundCmd = cmd
		}
	}

	// Command not found
	if foundCmd == nil {
		SendMessage(w, http.StatusNotFound, "Command not recognized")
		return
	}

	// Execute command and get response
	respdata, execErr := foundCmd.Callback(req.Args)
	if execErr != nil {
		SendMessage(w, http.StatusInternalServerError, "Unable to successfully execute command")
		return
	}

	// Send response to consumer
	SendData(w, http.StatusOK, map[string]string{"data": *respdata})
}

/*
Add a new function to the server
*/
func (srv *CmdServer) RegisterCmd(name string, callback CmdCallback) {
	cmd := Command{name, callback}
	srv.Commands = append(srv.Commands, &cmd)
}
