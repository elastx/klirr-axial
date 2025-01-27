package synchronization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"axial/api"
	"axial/models"
)

func SyncMessages(node models.RemoteNode, message []models.Message) error {

	syncMessagesRequest := api.SyncMessagesRequest{
		Messages: message,
	}

	jsonMessage, err := json.Marshal(syncMessagesRequest)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("http://%s/v1/sync/messages", node.Address)
	fmt.Printf("Sending message to %s: %s\n", endpoint, string(jsonMessage))
	response, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonMessage))
	if err != nil {
		return err
	}

	fmt.Printf("Received response from %s: %+v\n", node.Address, response)

	defer response.Body.Close()

	return nil
}
