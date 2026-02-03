package synchronization

import (
	"fmt"

	"axial/models"
	"axial/remote"
)

func SyncUsers(node remote.API, users []models.User) error {
	endpoint := node.SyncUsers()
	responseData, response, err := endpoint.Post(users)
	if err != nil {
		return err
	}

	fmt.Printf("Received %s response from %s: %+v\n", response.Status, node.Address, responseData)

	return nil
}
