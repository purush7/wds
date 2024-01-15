package util

import (
	"net/http"
	"time"
)

func GetTraceableHTTPClient(timeout *time.Duration, resourceName string) *http.Client {
	var client *http.Client
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 10
	t.MaxConnsPerHost = 10
	t.MaxIdleConnsPerHost = 10
	if timeout != nil {
		client = &http.Client{
			Timeout: *timeout,
		}
	} else {
		client = &http.Client{}
	}
	client.Transport = t

	return client
}
