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

	endpoint := fmt.Sprintf("http://%s/v1/messages", node.Address)
	fmt.Printf("Sending message to %s: %s\n", endpoint, string(jsonMessage))
	response, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonMessage))
	if err != nil {
		return err
	}

	fmt.Printf("Received response from %s: %+v\n", node.Address, response)

	defer response.Body.Close()

	return nil
}
