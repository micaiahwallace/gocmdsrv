package cmdserver

import (
	"encoding/json"
	"net/http"
	"os/exec"
)

/*
Parse API request body into struct
*/
func ParseExecuteRequest(body []byte) (*ExecuteApiRequest, error) {

	// create new request object
	req := ExecuteApiRequest{}

	// decode json
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	// return object
	return &req, nil
}

/*
Send an http status and { message: "message" } response body
*/
func SendMessage(w http.ResponseWriter, status int, message string) {
	SendData(w, status, map[string]string{"message": message})
}

/*
Send data back to the api consumer
*/
func SendData(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

/*
Execute a command with arguments
*/
func ExecWithArgs(cmd string, preArgs []string) func([]string) (*string, error) {

	return func(args []string) (*string, error) {

		// Prepare cli args
		allArgs := append(preArgs, args...)
		c := exec.Command(cmd, allArgs...)

		// start command execution
		response, err := c.CombinedOutput()

		// return command execution error
		if err != nil {
			return nil, err
		}

		// return string response
		resp := string(response)
		return &resp, nil
	}
}
