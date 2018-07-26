package app

import (
	"net/http"

	"google.golang.org/appengine"
)

// GAEEntityKind is for GAE Entity kind
const (
	GAEEntityKind = "GCEOperations"
)

func init() {
	http.HandleFunc("/cron/operations", operationHandle)
	http.HandleFunc("/cron/worker", workHandle)
	appengine.Main()
}
