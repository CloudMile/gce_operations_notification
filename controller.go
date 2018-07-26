package app

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
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

func operationHandle(w http.ResponseWriter, r *http.Request) {
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

func workHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cron/worker" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	ctx := appengine.NewContext(r)

	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: google.AppEngineTokenSource(ctx, compute.ComputeScope),
			Base: &urlfetch.Transport{
				Context: ctx,
			},
		},
	}

	computeService, err := compute.New(client)
	if err != nil {
		log.Errorf(ctx, "compute error: %s", err)
	}

	outputOperations := OutputOperations{
		Ctx:            ctx,
		ComputeService: computeService,
		Projects:       getProjectIDs(),
		Itmes:          make(map[string][]*OutputOperation),
	}
	outputOperations.GetGceUnkownAggregatedList()
	outputOperations.SendMail()
}

func getProjectIDs() (projectIDs []string) {
	projectIDs = strings.Split(os.Getenv("PROJECT_IDS"), `,`)
	return
}
