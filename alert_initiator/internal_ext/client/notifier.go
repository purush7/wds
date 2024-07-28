package client

import (
	"alert_system/util"
	"net/http"
	"sync"
)

var notifierOnce sync.Once
var notifierClient *http.Client

func GetNotfierClient() *http.Client {
	notifierOnce.Do(func() {
		notifierClient = util.GetTraceableHTTPClient(nil, "notifier")
	})
	return notifierClient
}
