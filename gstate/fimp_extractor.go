package gstate

import (
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/connector/plugins/fimpmqtt"
	"github.com/thingsplex/tpflow/flow/context"
	"github.com/thingsplex/tpflow/registry/storage"
	"runtime/debug"
	"strings"
	"time"
)

// FimpStateExtractor is responsible for capturing state from FIMP mqtt message stream.
type FimpStateExtractor struct {
	msgTransport *fimpgo.MqttTransport
	config       *fimpmqtt.ConnectorConfig
	contextDb    *context.Context
	registry     storage.RegistryStorage
	regCache     map[string]RegCacheRecord
}

func NewFimpStateExtractor(config *fimpmqtt.ConnectorConfig, contextDb *context.Context, registry storage.RegistryStorage) *FimpStateExtractor {
	regCache := make(map[string]RegCacheRecord)
	return &FimpStateExtractor{config: config, contextDb: contextDb, registry: registry, regCache: regCache}
}

func (mg *FimpStateExtractor) InitMessagingTransport() {
	clientId := mg.config.MqttClientIdPrefix + "state_tracker"
	mg.msgTransport = fimpgo.NewMqttTransport(mg.config.MqttServerURI, clientId, mg.config.MqttUsername, mg.config.MqttPassword, true, 1, 1)
	mg.msgTransport.SetGlobalTopicPrefix(mg.config.MqttTopicGlobalPrefix)
	err := mg.msgTransport.Start()
	log.Info("<gstate> Mqtt transport connected")
	if err != nil {
		log.Error("<gstate> Error connecting to broker : ", err)
	}
	mg.msgTransport.SetMessageHandler(mg.onMqttMessage)
	mg.msgTransport.Subscribe("pt:j1/mt:evt/rt:dev/#")

}

func (mg *FimpStateExtractor) onMqttMessage(topic string, addr *fimpgo.Address, iotMsg *fimpgo.FimpMessage, rawMessage []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("<gstate> State tracker crashed with error : ", string(debug.Stack()))
			log.Errorf("<gstate> Crashed while processing message from topic = %s ", topic)
			debug.PrintStack()
		}
	}()
	var externalId int
	intfSplit := strings.Split(iotMsg.Type, ".")
	stateName := ""
	if len(intfSplit) > 0 {
		stateName = intfSplit[1]
	}
	stateName = stateName + "@" + strings.Replace(topic, "pt:j1/mt:evt/", "", -1)

	cacheRec, cacheRecordExists := mg.regCache[topic]
	doLookup := true

	if cacheRecordExists {
		if time.Since(cacheRec.updatedAt) > time.Second*60 { // lookup ttl 1 minute
			log.Trace("<gstate> Cache record expired , doing lookup")
		} else {
			externalId = cacheRec.externalId
			doLookup = false
		}
	}

	if doLookup {
		svcs, err := mg.registry.GetAllServices(&storage.ServiceFilter{Topic: topic})
		if err == nil && len(svcs) > 0 {
			externalId = int(svcs[0].ParentContainerId)
			mg.regCache[topic] = RegCacheRecord{
				updatedAt:  time.Now(),
				externalId: externalId,
			}

		} else {
			log.Debugf("<gstate> Global state can't find device reference . Lookup topi topic = %s", topic)
		}
	}

	//TODO: Emit event if state has changed. Introduce trigger that can react on that event.

	mg.contextDb.SetState(stateName, iotMsg.ValueType, iotMsg.Value, externalId)
	log.Tracef("<gstate> State updated , name = %s", stateName)
}
