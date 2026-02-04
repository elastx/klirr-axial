package remote

import (
	"net/http"

	"axial/api"
	"axial/models"
)

func (n *API) Messages() Endpoint[models.CreateMessage, models.Message, models.Message] {
	return Endpoint[models.CreateMessage, models.Message, models.Message]{
		Node:    n,
		Version: "v1",
		Path:    "v1/messages",
		ValidPostResponse: func(response http.Response) bool {
			return response.StatusCode == http.StatusCreated
		},
		ValidGetResponse: func(response http.Response) bool {
			return response.StatusCode == http.StatusOK || response.StatusCode == http.StatusNotFound
		},
	}
}

func (n *API) Users() Endpoint[models.CreateUser, models.User, []models.User] {
	return Endpoint[models.CreateUser, models.User, []models.User]{
		Node:    n,
		Version: "v1",
		Path:    "users",
		ValidPostResponse: func(response http.Response) bool {
			return response.StatusCode == http.StatusOK
		},
	}
}

func (n *API) Ping() Endpoint[interface{}, interface{}, api.PingResponse] {
	return Endpoint[interface{}, interface{}, api.PingResponse]{
		Node:    n,
		Version: "v1",
		Path:    "ping",
	}
}

func (n *API) Sync() Endpoint[api.SyncRequest, api.SyncResponse, interface{}] {
	return Endpoint[api.SyncRequest, api.SyncResponse, interface{}]{
		Node:    n,
		Version: "v1",
		Path:    "sync",
	}
}

func (n *API) SyncMessages() Endpoint[[]models.Message, interface{}, interface{}] {
	return Endpoint[[]models.Message, interface{}, interface{}]{
		Node:    n,
		Version: "v1",
		Path:    "sync/messages",
	}
}

func (n *API) SyncBulletins() Endpoint[[]models.Bulletin, interface{}, interface{}] {
	return Endpoint[[]models.Bulletin, interface{}, interface{}]{
		Node:    n,
		Version: "v1",
		Path:    "sync/bulletins",
	}
}

func (n *API) SyncUsers() Endpoint[[]models.User, interface{}, interface{}] {
	return Endpoint[[]models.User, interface{}, interface{}]{
		Node:    n,
		Version: "v1",
		Path:    "sync/users",
	}
}
