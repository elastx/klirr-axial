package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type API struct {
	Scheme  string
	Address string
	Port    int
}

type Endpoint[PostRequestType any, PostResponseType any, GetResponseType any] struct {
	Node              *API
	Version           string
	Path              string
	ValidPostResponse func(http.Response) bool
	ValidGetResponse  func(http.Response) bool
}

func (e *Endpoint[_, _, _]) URL() *url.URL {
	if e.Node == nil {
		return nil
	}
	return &url.URL{
		Scheme: e.Node.Scheme,
		Host:   fmt.Sprintf("%s:%d", e.Node.Address, e.Node.Port),
		Path:   e.Path,
	}
}

func (e *Endpoint[T, R, _]) Post(data T) (R, *http.Response, error) {
	var result R
	body, err := json.Marshal(data)
	if err != nil {
		return result, nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	resp, err := http.Post(e.URL().String(), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return result, resp, fmt.Errorf("failed to perform POST request: %w", err)
	}

	if e.ValidPostResponse != nil && !e.ValidPostResponse(*resp) {
		return result, resp, fmt.Errorf("invalid response: %s", resp.Status)
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, resp, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, resp, nil
}

func (e *Endpoint[_, _, R]) Get() (R, *http.Response, error) {
	var result R
	resp, err := http.Get(e.URL().String())
	if err != nil {
		return result, resp, fmt.Errorf("failed to perform GET request: %w", err)
	}

	if e.ValidGetResponse != nil && !e.ValidGetResponse(*resp) {
		return result, resp, fmt.Errorf("invalid response: %s", resp.Status)
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, resp, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, resp, nil
}

func ExampleCode() {
	node := &API{
		Scheme:  "http",
		Address: "localhost",
		Port:    8080,
	}

	users := node.Users()
	user, response, err := users.Get()
	if err != nil {
		fmt.Println("Error fetching users:", err)
		return
	}

	fmt.Println("Received", response.Status, "response from", node.Address)
	
	fmt.Println(user)
}
