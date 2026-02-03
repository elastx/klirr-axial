package synchronization

import (
	"fmt"

	"axial/api"
	"axial/models"
	"axial/remote"
)

func SyncMessages(node remote.API, message []models.Message) error {
	endpoint := node.SyncMessages()
	responseData, response, err := endpoint.Post(api.SyncMessagesRequest{Messages: message})
	if err != nil {
		return err
	}

	fmt.Printf("Received %s response from %s: %+v\n", response.Status, node.Address, responseData)

	return nil
}
