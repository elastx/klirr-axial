package synchronization

import (
	"fmt"

	"axial/api"
	"axial/models"
	"axial/remote"
)

func SyncBulletin(node remote.API, posts []models.Bulletin) error {
	endpoint := node.SyncBulletin()
	responseData, response, err := endpoint.Post(api.SyncBulletinRequest{Posts: posts})
	if err != nil {
		return err
	}

	fmt.Printf("Received %s response from %s: %+v\n", response.Status, node.Address, responseData)

	return nil
}
