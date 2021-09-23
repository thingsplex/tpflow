package gstate

import (
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow"
	"github.com/thingsplex/tpflow/flow/context"
	"runtime/debug"
	"strings"
)

type GlobalStateTracker struct {
	fimpStateExtractor *FimpStateExtractor
	config *tpflow.Configs
	contextDb * context.Context
}

func NewGlobalStateTracker(config *tpflow.Configs, contextDb *context.Context) *GlobalStateTracker {
	return &GlobalStateTracker{config: config, contextDb: contextDb,fimpStateExtractor: NewFimpStateExtractor(config,contextDb)}
}

func (gs *GlobalStateTracker) Init() {
	gs.fimpStateExtractor.InitMessagingTransport()
}


type FimpStateExtractor struct {
	msgTransport *fimpgo.MqttTransport
	config *tpflow.Configs
	contextDb * context.Context
}

func NewFimpStateExtractor(config *tpflow.Configs, contextDb *context.Context) *FimpStateExtractor {
	return &FimpStateExtractor{config: config, contextDb: contextDb}
}

func (mg *FimpStateExtractor) InitMessagingTransport() {
	clientId := mg.config.MqttClientIdPrefix + "state_tracker"
	mg.msgTransport = fimpgo.NewMqttTransport(mg.config.MqttServerURI, clientId, mg.config.MqttUsername, mg.config.MqttPassword, true, 1, 1)
	mg.msgTransport.SetGlobalTopicPrefix(mg.config.MqttTopicGlobalPrefix)
	err := mg.msgTransport.Start()
	log.Info("<MqRegInt> Mqtt transport connected")
	if err != nil {
		log.Error("<MqRegInt> Error connecting to broker : ", err)
	}
	mg.msgTransport.SetMessageHandler(mg.onMqttMessage)
	mg.msgTransport.Subscribe("pt:j1/mt:evt/rt:dev/#")

}

func (mg *FimpStateExtractor) onMqttMessage(topic string, addr *fimpgo.Address, iotMsg *fimpgo.FimpMessage, rawMessage []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("<gstate> State tracker crashed with error : ", r)
			log.Errorf("<MqRegInt> Crashed while processing message from topic = %s msgType = %s", r, addr.MsgType)
			debug.PrintStack()
		}
	}()
	intfSplit := strings.Split(iotMsg.Type,".")
	stateName := ""
	if len(intfSplit)>0 {
		stateName = intfSplit[1]
	}
	stateName = stateName + "@" + strings.Replace(topic,"pt:j1/mt:evt/","",-1)

	//TODO: Emit event if state has changed.

	mg.contextDb.SetState(stateName,iotMsg.ValueType,iotMsg.Value,0)
	log.Tracef("<gstate> State updated , name = %s",stateName)
}