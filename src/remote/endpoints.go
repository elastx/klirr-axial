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

func (n *API) SyncMessages() Endpoint[api.SyncMessagesRequest, interface{}, interface{}] {
	return Endpoint[api.SyncMessagesRequest, interface{}, interface{}]{
		Node:    n,
		Version: "v1",
		Path:    "sync/messages",
		ValidPostResponse: func(response http.Response) bool {
			return response.StatusCode == http.StatusCreated
		},
	}
}

func (n *API) SyncUsers() Endpoint[api.SyncUsersRequest, interface{}, interface{}] {
	return Endpoint[api.SyncUsersRequest, interface{}, interface{}]{
		Node:    n,
		Version: "v1",
		Path:    "sync/users",
		ValidPostResponse: func(response http.Response) bool {
			return response.StatusCode == http.StatusCreated
		},
	}
}

func (n *API) SyncFiles() Endpoint[api.SyncFilesRequest, interface{}, interface{}] {
	return Endpoint[api.SyncFilesRequest, interface{}, interface{}]{
		Node:    n,
		Version: "v1",
		Path:    "sync/files",
		ValidPostResponse: func(response http.Response) bool {
			return response.StatusCode == http.StatusCreated
		},
	}
}

func (n *API) SyncBulletin() Endpoint[api.SyncBulletinRequest, interface{}, interface{}] {
	return Endpoint[api.SyncBulletinRequest, interface{}, interface{}]{
		Node:    n,
		Version: "v1",
		Path:    "sync/bulletin",
		ValidPostResponse: func(response http.Response) bool {
			return response.StatusCode == http.StatusCreated
		},
	}
}

func (n *API) FileMetadata(fileID string) Endpoint[interface{}, interface{}, models.File] {
	return Endpoint[interface{}, interface{}, models.File]{
		Node:    n,
		Version: "v1",
		Path:    "files/metadata/" + fileID,
	}
}
