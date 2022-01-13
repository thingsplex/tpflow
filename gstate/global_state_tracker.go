package gstate

import (
	"github.com/thingsplex/tpflow/connector/plugins/fimpmqtt"
	"github.com/thingsplex/tpflow/flow/context"
	"github.com/thingsplex/tpflow/registry/storage"
	"time"
)

// GlobalStateTracker global state extractor is responsible for capturing states from multiple sources.
type GlobalStateTracker struct {
	fimpStateExtractor *FimpStateExtractor
	contextDb          *context.Context
	registry           storage.RegistryStorage
}

func NewGlobalStateTracker(contextDb *context.Context, registry storage.RegistryStorage) *GlobalStateTracker {
	return &GlobalStateTracker{contextDb: contextDb, registry: registry}
}

func (gs *GlobalStateTracker) InitFimpStateExtractor(config *fimpmqtt.ConnectorConfig) {
	gs.fimpStateExtractor = NewFimpStateExtractor(config, gs.contextDb, gs.registry)
	gs.fimpStateExtractor.InitMessagingTransport()
}

type RegCacheRecord struct {
	updatedAt  time.Time
	externalId int
}
