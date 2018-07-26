package app

import (
	"context"
	"os"
	"strconv"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// OutputOperation for store operation
type OutputOperation struct {
	OperationType string
	TargetLink    string
	StartTime     string
	EndTime       string
}

// OutputOperations for store list OutputOperation
type OutputOperations struct {
	Ctx            context.Context
	Itmes          map[string][]*OutputOperation
	ComputeService *compute.Service
	Projects       []string
}

// GetGceUnkownAggregatedList get AggregatedList from API
func (outO *OutputOperations) GetGceUnkownAggregatedList() {
	for _, project := range outO.Projects {
		aggregatedListTemp := gceUnkownAggregatedList(outO.Ctx, outO.ComputeService, project)

		for key, value := range aggregatedListTemp {
			log.Infof(outO.Ctx, "key: %s", key)
			outO.Itmes[key] = append(outO.Itmes[key], value...)
		}
	}
}

// SendMail is send mail
func (outO *OutputOperations) SendMail() {
	for operationType, operationList := range outO.Itmes {
		if len(operationList) > 0 {
			mail := Mail{
				Ctx:           outO.Ctx,
				OperationType: operationType,
				Operations:    operationList,
			}
			mail.Send()
		}
	}
}

func gceUnkownAggregatedList(ctx context.Context, computeService *compute.Service, project string) (unkownAggregatedList map[string][]*OutputOperation) {
	unkownAggregatedListTemp := make(map[string][]*OutputOperation)
	req := computeService.GlobalOperations.AggregatedList(project)
	req = req.Filter(os.Getenv("OPERATION_FILTER"))

	if err := req.Pages(ctx, func(page *compute.OperationAggregatedList) error {
		for _, operationsScopedList := range page.Items {
			if len(operationsScopedList.Operations) > 0 {
				for _, operation := range operationsScopedList.Operations {
					log.Infof(ctx, "id: %d, target %s", operation.Id, operation.TargetLink)
					isNew, outputOperation := getOrPutDatastore(ctx, operation)
					if isNew {
						unkownAggregatedListTemp[operation.OperationType] = append(unkownAggregatedListTemp[operation.OperationType], &outputOperation)
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

func getOrPutDatastore(ctx context.Context, operation *compute.Operation) (newDatastore bool, datastoreObject OutputOperation) {
	newDatastore = false
	datastoreKey := datastore.NewKey(ctx, GAEEntityKind, strconv.FormatUint(operation.Id, 10), 0, nil)

	if err := datastore.Get(ctx, datastoreKey, &datastoreObject); err != nil {
		log.Errorf(ctx, "datastore %s", err)
		if err.Error() == "datastore: no such entity" {
			datastoreObject.TargetLink = operation.TargetLink
			datastoreObject.OperationType = operation.OperationType
			datastoreObject.StartTime = operation.StartTime
			datastoreObject.EndTime = operation.EndTime
			if _, putErr := datastore.Put(ctx, datastoreKey, &datastoreObject); putErr != nil {
				log.Errorf(ctx, "datastore %s", putErr)
			}
			newDatastore = true
		}
	}

	return
}