package synchronization

import (
	"net/http"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var Client HTTPClient = &http.Client{Timeout: 15 * time.Second}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

var DefaultClock Clock = SystemClock{}
