package synchronization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"axial/models"
)

func SendMessage(node models.RemoteNode, message models.Message) error {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	fmt.Printf("Sending message to %s: %s\n", node.Address, string(jsonMessage))
	response, err := http.Post(fmt.Sprintf("http://%s/v1/message", node.Address), "application/json", bytes.NewBuffer(jsonMessage))
	if err != nil {
		return err
	}

	fmt.Printf("Received response from %s: %+v\n", node.Address, response)

	defer response.Body.Close()

	return nil
}
