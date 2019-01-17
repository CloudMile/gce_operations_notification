package main

import (
	"net/http"

	"github.com/CloudMile/gce_operations_notification/controller"
	"google.golang.org/appengine"
)

func main() {
	http.HandleFunc("/cron/operations", controller.OperationHandle)
	http.HandleFunc("/cron/worker", controller.WorkHandle)
	appengine.Main()
}
