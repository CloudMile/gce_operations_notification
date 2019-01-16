package model

import (
	"context"
	"strconv"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

// GAEEntityKind is for GAE Entity kind
const (
	GAEEntityKind = "GCEOperations"
)

type database interface {
	getOrPut() (bool, OutputOperation)
}

// GCPDatastore is struct for GCP datastore
type GCPDatastore struct {
	Ctx       context.Context
	Operation *compute.Operation
}

// GAEMemcache is struct for GAE memcache
type GAEMemcache struct {
	Ctx       context.Context
	Operation *compute.Operation
}

func (gcpDS GCPDatastore) getOrPut() (newDatastore bool, datastoreObject OutputOperation) {
	newDatastore = false
	datastoreKey := datastore.NewKey(gcpDS.Ctx, GAEEntityKind, strconv.FormatUint(gcpDS.Operation.Id, 10), 0, nil)

	if err := datastore.Get(gcpDS.Ctx, datastoreKey, &datastoreObject); err != nil {
		log.Warningf(gcpDS.Ctx, "datastore %s", err)
		if err == datastore.ErrNoSuchEntity {
			datastoreObject.TargetLink = gcpDS.Operation.TargetLink
			datastoreObject.OperationType = gcpDS.Operation.OperationType
			datastoreObject.StartTime = gcpDS.Operation.StartTime
			datastoreObject.EndTime = gcpDS.Operation.EndTime
			if _, putErr := datastore.Put(gcpDS.Ctx, datastoreKey, &datastoreObject); putErr != nil {
				log.Errorf(gcpDS.Ctx, "datastore %s", putErr)
			}
			newDatastore = true
		}
	}
	return
}

func (gaeMC GAEMemcache) getOrPut() (newMemcache bool, datastoreObject OutputOperation) {
	newMemcache = false
	memcacheKey := `GCE_` + strconv.FormatUint(gaeMC.Operation.Id, 10)

	_, err := memcache.JSON.Get(gaeMC.Ctx, memcacheKey, &datastoreObject)
	if err == memcache.ErrCacheMiss {
		datastoreObject.TargetLink = gaeMC.Operation.TargetLink
		datastoreObject.OperationType = gaeMC.Operation.OperationType
		datastoreObject.StartTime = gaeMC.Operation.StartTime
		datastoreObject.EndTime = gaeMC.Operation.EndTime

		item := &memcache.Item{
			Key:    memcacheKey,
			Object: datastoreObject,
		}

		if err := memcache.JSON.Set(gaeMC.Ctx, item); err != nil {
			log.Errorf(gaeMC.Ctx, "error adding item: %v", err)
		}
		newMemcache = true
	}
	return
}
