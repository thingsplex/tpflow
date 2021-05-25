package context

import (
	"github.com/pkg/errors"
	"sync"
)

type ContextInMemoryStore struct {
	store sync.Map
}

func (ctx *ContextInMemoryStore) Store(rec ContextRecord, flowId string) {
	inMemoryRecI, ok := ctx.store.Load(flowId)
	var inMemoryRec sync.Map
	if ok {
		inMemoryRec = inMemoryRecI.(sync.Map)
	}
	inMemoryRec.Store(rec.Name, rec)
	ctx.store.Store(flowId, inMemoryRec)
}

func (ctx *ContextInMemoryStore) Get(flowId string, varName string) *ContextRecord {
	flowI, ok := ctx.store.Load(flowId)
	if ok {
		flowS, ok := flowI.(sync.Map)
		if !ok {
			return nil
		}
		recI, ok := flowS.Load(varName)
		if !ok {
			return nil
		}
		rec, ok := recI.(ContextRecord)
		if ok {
			return &rec
		}
	}
	return nil
}

func (ctx *ContextInMemoryStore) DeleteFlow(flowId string) {
	ctx.store.Delete(flowId)
}

func (ctx *ContextInMemoryStore) GetRecordsForFlow(flowId string) ([]ContextRecord, error) {
	var result []ContextRecord
	flowI, ok := ctx.store.Load(flowId)
	if ok {
		flowS, ok := flowI.(sync.Map)
		if !ok {
			return result, errors.New("Wrong type")
		}
		flowS.Range(func(key, value interface{}) bool {
			rec, ok := value.(ContextRecord)
			if ok {
				result = append(result, rec)
			}
			return true
		})
	} else {
		//log.Debugf("---- Can't load records for for flowId = %s",flowId)
	}
	return result, nil
}
