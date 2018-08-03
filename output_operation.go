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
	var c = make(chan map[string][]*OutputOperation, len(outO.Projects))

	for _, project := range outO.Projects {
		go goGetOperation(outO, project, c)
	}

	for range outO.Projects {
		for key, value := range <-c {
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
	unkownAggregatedList = make(map[string][]*OutputOperation)
	req := computeService.GlobalOperations.AggregatedList(project)
	req = req.Filter(os.Getenv("OPERATION_FILTER"))

	if err := req.Pages(ctx, func(page *compute.OperationAggregatedList) error {
		var c = make(chan map[string][]*OutputOperation, len(page.Items))

		for _, operationsScopedList := range page.Items {
			go goGceUnkownAggregatedList(ctx, operationsScopedList, c)
		}

		for range page.Items {
			items := <-c
			if len(items) > 0 {
				for key, value := range items {
					unkownAggregatedList[key] = append(unkownAggregatedList[key], value...)
				}
			}
		}

		return nil
	}); err != nil {
		log.Errorf(ctx, "compute error: %s", err)
	}
	return
}

func multiCheckOperation(ctx context.Context, arrayLen int, operationsScopedList compute.OperationsScopedList) (unkownAggregatedList map[string][]*OutputOperation) {
	unkownAggregatedList = make(map[string][]*OutputOperation)
	var c = make(chan map[string][]*OutputOperation, arrayLen)

	for _, operation := range operationsScopedList.Operations {
		go goMultiCheckOperation(ctx, operation, c)
	}

	for range operationsScopedList.Operations {
		items := <-c
		if len(items) > 0 {
			for key, value := range items {
				unkownAggregatedList[key] = append(unkownAggregatedList[key], value...)
			}
		}
	}
	return
}

func goGetOperation(outO *OutputOperations, project string, c chan map[string][]*OutputOperation) {
	aggregatedListTemp := gceUnkownAggregatedList(outO.Ctx, outO.ComputeService, project)
	items := make(map[string][]*OutputOperation)

	for key, value := range aggregatedListTemp {
		log.Infof(outO.Ctx, "key: %s", key)
		items[key] = append(items[key], value...)
	}
	c <- items
	return
}

func goMultiCheckOperation(ctx context.Context, operation *compute.Operation, c chan map[string][]*OutputOperation) {
	items := make(map[string][]*OutputOperation)
	log.Infof(ctx, "id: %d, target %s", operation.Id, operation.TargetLink)
	isNew, outputOperation := getOrPut(ctx, operation)
	if isNew {
		items[operation.OperationType] = append(items[operation.OperationType], &outputOperation)
	}
	c <- items
}

func goGceUnkownAggregatedList(ctx context.Context, operationsScopedList compute.OperationsScopedList, c chan map[string][]*OutputOperation) {
	items := make(map[string][]*OutputOperation)
	lenOperationsScopedListOperations := len(operationsScopedList.Operations)

	if lenOperationsScopedListOperations > 0 {
		unkownAggregatedListTemp := multiCheckOperation(ctx, lenOperationsScopedListOperations, operationsScopedList)
		for key, value := range unkownAggregatedListTemp {
			items[key] = append(items[key], value...)
		}
	}
	c <- items
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
