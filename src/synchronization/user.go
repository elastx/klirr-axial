package synchronization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"axial/api"
	"axial/models"
)

func SyncUsers(node models.RemoteNode, users []models.User) error {
	syncUsersRequest := api.SyncUsersRequest{
		Users: users,
	}

	jsonRegistration, err := json.Marshal(syncUsersRequest)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("http://%s/v1/sync/users", node.Address)
	fmt.Printf("Sending user registration to %s: %s\n", endpoint, string(jsonRegistration))
	response, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonRegistration))
	if err != nil {
		return err
	}

	defer response.Body.Close()

	return nil
}
