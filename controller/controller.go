package controller

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/CloudMile/gce_operations_notification/model"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	switch status {
	case http.StatusNotFound:
		fmt.Fprint(w, "404 Not Found")
	case http.StatusMethodNotAllowed:
		fmt.Fprint(w, "405 Method Not Allow")
	default:
		fmt.Fprint(w, "Bad Request")
	}
}

// OperationHandle is GET /cron/operations
func OperationHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cron/operations" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	ctx := appengine.NewContext(r)
	log.Infof(ctx, "query: %v", r.URL.Query())

	t := taskqueue.NewPOSTTask("/cron/worker", r.URL.Query())
	if _, err := taskqueue.Add(ctx, t, "get-operations"); err != nil {
		errorHandler(w, r, http.StatusInternalServerError)
		return
	}
}

// WorkHandle is POST /cron/worker for queue
func WorkHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cron/worker" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	ctx := appengine.NewContext(r)
	cs := model.ComputeService{Ctx: ctx}
	cs.Get()
	if cs.Error != nil {
		log.Errorf(ctx, "compute error: %s", cs.Error)
		return
	}

	outputOperations := model.OutputOperations{
		Ctx:            ctx,
		ComputeService: cs.ComputeService,
		Projects:       getProjectIDs(),
		Itmes:          make(map[string][]*model.OutputOperation),
	}
	outputOperations.GetGceUnkownAggregatedList()
	outputOperations.SendMail()
}

func getProjectIDs() (projectIDs []string) {
	projectIDs = strings.Split(os.Getenv("PROJECT_IDS"), `,`)
	return
}
