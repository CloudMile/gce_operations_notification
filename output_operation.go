package app

import (
	"context"
	"os"

	compute "google.golang.org/api/compute/v1"
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
	var c = make(chan bool, len(outO.Projects))

	for _, project := range outO.Projects {
		go func(project string) {
			aggregatedListTemp := gceUnkownAggregatedList(outO.Ctx, outO.ComputeService, project)

			for key, value := range aggregatedListTemp {
				log.Infof(outO.Ctx, "key: %s", key)
				outO.Itmes[key] = append(outO.Itmes[key], value...)
			}
			c <- true
		}(project)
	}

	for range outO.Projects {
		<-c
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
	unkownAggregatedList = make(map[string][]*OutputOperation)
	req := computeService.GlobalOperations.AggregatedList(project)
	req = req.Filter(os.Getenv("OPERATION_FILTER"))

	if err := req.Pages(ctx, func(page *compute.OperationAggregatedList) error {
		var c = make(chan bool, len(page.Items))

		for _, operationsScopedList := range page.Items {
			go func(operationsScopedList compute.OperationsScopedList) {
				lenOperationsScopedListOperations := len(operationsScopedList.Operations)

				if lenOperationsScopedListOperations > 0 {
					unkownAggregatedList = multiCheckOperation(ctx, lenOperationsScopedListOperations, operationsScopedList)
				}
				c <- true
			}(operationsScopedList)
		}

		for range page.Items {
			<-c
		}

		return nil
	}); err != nil {
		log.Errorf(ctx, "compute error: %s", err)
	}
	return
}

func multiCheckOperation(ctx context.Context, arrayLen int, operationsScopedList compute.OperationsScopedList) (unkownAggregatedList map[string][]*OutputOperation) {
	unkownAggregatedList = make(map[string][]*OutputOperation)
	var c = make(chan bool, arrayLen)

	for _, operation := range operationsScopedList.Operations {
		go func(operation *compute.Operation) {
			log.Infof(ctx, "id: %d, target %s", operation.Id, operation.TargetLink)
			isNew, outputOperation := getOrPut(ctx, operation)
			if isNew {
				unkownAggregatedList[operation.OperationType] = append(unkownAggregatedList[operation.OperationType], &outputOperation)
			}
			c <- true
		}(operation)
	}
	for range operationsScopedList.Operations {
		<-c
	}
	return
}

func getOrPut(ctx context.Context, operation *compute.Operation) (bool, OutputOperation) {
	dbMap := map[string]database{
		"datastore": GCPDatastore{Ctx: ctx, Operation: operation},
		"memcache":  GAEMemcache{Ctx: ctx, Operation: operation},
		"":          GCPDatastore{Ctx: ctx, Operation: operation},
	}

	databaseType := ""
	for _, dbType := range []string{"datastore", "memcache"} {
		if os.Getenv(`DATABASE`) == dbType {
			databaseType = dbType
			break
		}
	}

	return dbMap[databaseType].getOrPut()
}
