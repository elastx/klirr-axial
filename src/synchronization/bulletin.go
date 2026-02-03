package synchronization

import (
	"fmt"

	"axial/models"
	"axial/remote"
)

func SyncBulletins(node remote.API, bulletins []models.Bulletin) error {
	endpoint := node.SyncBulletins()
	responseData, response, err := endpoint.Post(bulletins)
	if err != nil {
		return err
	}

	fmt.Printf("Received %s response from %s: %+v\n", response.Status, node.Address, responseData)

	return nil
}
