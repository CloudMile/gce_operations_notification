package app

import (
	"net/http"

	"google.golang.org/appengine"
)

func init() {
	http.HandleFunc("/cron/operations", operationHandle)
	http.HandleFunc("/cron/worker", workHandle)
	appengine.Main()
}
