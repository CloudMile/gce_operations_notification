package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

// GAEEntityKind is for GAE Entity kind
const (
	GAEEntityKind = "GCEOperations"
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

	projects := getProjectIDs()
	aggregatedList := make(map[string][]*compute.Operation)

	for _, project := range projects {
		aggregatedListTemp := gceUnkownAggregatedList(ctx, computeService, project)
		for key, value := range aggregatedListTemp {
			aggregatedList[key] = append(aggregatedList[key], value...)
		}
	}

	for operationType, operationList := range aggregatedList {
		if len(operationList) > 0 {
			mail := Mail{
				Ctx:           ctx,
				OperationType: operationType,
				Operations:    operationList,
			}
			mail.Send()
		}
	}
}

func gceUnkownAggregatedList(ctx context.Context, computeService *compute.Service, project string) (unkownAggregatedList map[string][]*compute.Operation) {
	unkownAggregatedListTemp := make(map[string][]*compute.Operation)
	req := computeService.GlobalOperations.AggregatedList(project)
	req = req.Filter(os.Getenv("OPERATION_FILTER"))

	if err := req.Pages(ctx, func(page *compute.OperationAggregatedList) error {
		for _, operationsScopedList := range page.Items {
			if len(operationsScopedList.Operations) > 0 {
				for _, operation := range operationsScopedList.Operations {
					log.Infof(ctx, "id: %d, target %s", operation.Id, operation.TargetLink)
					if getOrPutDatastore(ctx, operation) {
						unkownAggregatedListTemp[operation.OperationType] = append(unkownAggregatedListTemp[operation.OperationType], operation)
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Errorf(ctx, "compute error: %s", err)
	}
	return unkownAggregatedListTemp
}

func getOrPutDatastore(ctx context.Context, operation *compute.Operation) (newDatastore bool) {
	newDatastore = false
	datastoreObject := OutputOperation{}
	datastoreKey := datastore.NewKey(ctx, GAEEntityKind, strconv.FormatUint(operation.Id, 10), 0, nil)

	if err := datastore.Get(ctx, datastoreKey, &datastoreObject); err != nil {
		log.Errorf(ctx, "datastore %s", err)
		if err.Error() == "datastore: no such entity" {
			putDatasotreObject := OutputOperation{
				TargetLink:    operation.TargetLink,
				OperationType: operation.OperationType,
			}
			if _, putErr := datastore.Put(ctx, datastoreKey, &putDatasotreObject); putErr != nil {
				log.Errorf(ctx, "datastore %s", putErr)
			}
			newDatastore = true
		}
	}

	return
}

func getProjectIDs() (projectIDs []string) {
	projectIDs = strings.Split(os.Getenv("PROJECT_IDS"), `,`)
	return
}
