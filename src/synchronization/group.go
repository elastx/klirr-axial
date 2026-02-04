package synchronization

import (
	"fmt"

	"axial/models"
	"axial/remote"
)

func SyncGroups(node remote.API, groups []models.Group) error {
	endpoint := node.SyncGroups()
	responseData, response, err := endpoint.Post(groups)
	if err != nil {
		return err
	}

	fmt.Printf("Received %s response from %s: %+v\n", response.Status, node.Address, responseData)

	return nil
}
