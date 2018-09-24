package flow

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/alivinco/fimpgo"
	tpflow "github.com/alivinco/tpflow"
	"github.com/alivinco/tpflow/connector"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/node"
	"github.com/alivinco/tpflow/utils"
	//thingsplexmodel "github.com/alivinco/thingsplex/model"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Manager struct {
	flowRegistry      []*Flow
	msgStreams        map[string]model.MsgPipeline
	msgTransport      *fimpgo.MqttTransport
	globalContext     *model.Context
	config            tpflow.Configs
	connectorRegistry connector.Registry
}



type FlowListItem struct {
	Id string
	Name string
	Group string
	Description string
	State string
	TriggerCounter int64
	ErrorCounter int64
	Stats *model.FlowStatsReport
}

func NewManager(config tpflow.Configs) (*Manager,error) {
	var err error
	man := Manager{config:config}
	man.msgStreams = make(map[string]model.MsgPipeline)
	man.flowRegistry = make([]*Flow,0)
	man.globalContext,err = model.NewContextDB(config.ContextStorageDir)
	man.globalContext.RegisterFlow("global")
	man.connectorRegistry = *connector.NewRegistry(config.ConnectorStorageDir)
	man.connectorRegistry.LoadInstancesFromDisk()
	return &man,err
}

func (mg *Manager) InitMessagingTransport() {
	clientId := mg.config.MqttClientIdPrefix+"flow_manager"
	mg.msgTransport = fimpgo.NewMqttTransport(mg.config.MqttServerURI, clientId, mg.config.MqttUsername, mg.config.MqttPassword, true, 1, 1)
	mg.msgTransport.SetGlobalTopicPrefix(mg.config.MqttTopicGlobalPrefix)
	err := mg.msgTransport.Start()
	log.Info("<FlMan> Mqtt transport connected")
	if err != nil {
		log.Error("<FlMan> Error connecting to broker : ", err)
	}
	mg.msgTransport.SetMessageHandler(mg.onMqttMessage)

}

func (mg *Manager) onMqttMessage(topic string, addr *fimpgo.Address, iotMsg *fimpgo.FimpMessage, rawMessage []byte) {
	msg := model.Message{AddressStr: topic, Address: *addr, Payload: *iotMsg,RawPayload:rawMessage}
	// Message broadcast to all flows
	for id, stream := range mg.msgStreams {
		if mg.GetFlowById(id).IsFlowInterestedInMessage(topic) {
			select {
			case stream <- msg:
				log.Debug("<FlMan> Message was sent to flow with id = ", id)
			default:
				log.Debug("<FlMan> Message is dropped (no listeners) for flow with id = ", id)
			}
		}
	}
}

func (mg *Manager) GenerateNewFlow() model.FlowMeta {
	fl := model.FlowMeta{}
	fl.Nodes = []model.MetaNode{{Id:"1",Type:"trigger",Label:"no label",Config:node.TriggerConfig{Timeout:0,ValueFilter:model.Variable{},IsValueFilterEnabled:false}}}
	fl.Id = utils.GenerateId(15)
	return fl
}

func (mg *Manager) GetNewStream(Id string) model.MsgPipeline {
	msgStream := make(model.MsgPipeline,10)
	mg.msgStreams[Id] = msgStream
	return msgStream
}

func (mg *Manager) GetGlobalContext() *model.Context {
	return mg.globalContext
}

func (mg *Manager) GetFlowFileNameById(id string ) string {
	return filepath.Join(mg.config.FlowStorageDir,id+".json")
}

func (mg *Manager) LoadAllFlowsFromStorage () error {
	files, err := ioutil.ReadDir(mg.config.FlowStorageDir)
	if err != nil {
		log.Error(err)
		return err
	}
	for _, file := range files {
		if strings.Contains(file.Name(),".json"){
			mg.LoadFlowFromFile(filepath.Join(mg.config.FlowStorageDir,file.Name()))
			mg.StartFlow(strings.Replace(file.Name(),".json","",1))
		}
	}
	return nil
}

func (mg *Manager) LoadFlowFromFile(fileName string) error{
	log.Info("<FlMan> Loading flow from file : ",fileName)
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Error("<FlMan> Can't open Flow file.")
		return err
	}
	mg.LoadFlowFromJson(file)
	return nil
}

func (mg *Manager) LoadFlowFromJson(flowJsonDef []byte) error{
	flowMeta := model.FlowMeta{}
	err := json.Unmarshal(flowJsonDef, &flowMeta)
	if err != nil {
		log.Error("<FlMan> Can't unmarshel DB file.")
		return err
	}

	flow := NewFlow(flowMeta, mg.globalContext, mg.msgTransport)
	flow.SetStoragePath(mg.config.FlowStorageDir)
	flow.SetConnectorRegistry(&mg.connectorRegistry)
	mg.flowRegistry = append(mg.flowRegistry,flow)
	return nil
}



func (mg *Manager) UpdateFlowFromJson(id string, flowJsonDef []byte) error {
	mg.StopFlow(id)
	mg.DeleteFlowFromRegistry(id,false)
	err := mg.LoadFlowFromJson(flowJsonDef)
	return err
}

func (mg *Manager) ReloadFlowFromStorage(id string ) error {
	mg.StopFlow(id)
	return mg.LoadFlowFromFile(mg.GetFlowFileNameById(id))
}

func (mg *Manager) ImportFlow(flowJsonDef []byte) (error) {
	newId := utils.GenerateId(15)
	flowMeta := model.FlowMeta{}
	err := json.Unmarshal(flowJsonDef, &flowMeta)
	if err != nil {
		log.Error("<FlMan> Can't unmarshel imported flow.")
		return err
	}
	oldId := flowMeta.Id
	flowAsString := string(flowJsonDef)
	flowAsStringUpdated := strings.Replace(flowAsString,oldId,newId,-1)
	fileName := mg.GetFlowFileNameById(newId)
	err = ioutil.WriteFile(fileName, []byte(flowAsStringUpdated), 0644)
	if err != nil {
		log.Error("Can't save flow to file . Error : ",err)
		return err
	}
	log.Debugf("<FlMan> Flow is imported")
	return mg.LoadFlowFromFile(fileName)
}

//
func (mg *Manager) UpdateFlowFromJsonAndSaveToStorage(id string, flowJsonDef []byte) (error) {
	fileName := mg.GetFlowFileNameById(id)
	log.Debugf("<FlMan> Saving flow to file %s , data size %d :",fileName,len(flowJsonDef))

	err := ioutil.WriteFile(fileName, flowJsonDef, 0644)
	if err != nil {
		log.Error("Can't save flow to file . Error : ",err)
		return err
	}
	err = mg.UpdateFlowFromJson(id,flowJsonDef)
	if err == nil {
		mg.StartFlow(id)
	}
	return err

}

func (mg *Manager) GetFlowById(id string) *Flow{
	for i := range mg.flowRegistry {
		if mg.flowRegistry[i].Id == id {
			return mg.flowRegistry[i]
		}
	}
	return nil
}

func (mg *Manager) GetFlowList() []FlowListItem{
	response := make ([]FlowListItem,len(mg.flowRegistry))
	var c int
	for i := range mg.flowRegistry {
		response[c] = FlowListItem{
			Id:mg.flowRegistry[i].Id,
			Name:mg.flowRegistry[i].Name,
			Group:mg.flowRegistry[i].FlowMeta.Group,
			Description:mg.flowRegistry[i].Description,
			TriggerCounter:mg.flowRegistry[i].TriggerCounter,
			ErrorCounter:mg.flowRegistry[i].ErrorCounter,
			State:mg.flowRegistry[i].opContext.State,
			Stats:mg.flowRegistry[i].GetFlowStats()}
		c++
	}
	return response
}

func (mg *Manager) ControlFlow(cmd string , flowId string) error {
	switch cmd {
	case "START":
	    mg.StartFlow(flowId)
	case "STOP":
		mg.StopFlow(flowId)
	}
	return nil
}

func (mg *Manager) StartFlow(flowId string) {
	flow := mg.GetFlowById(flowId)
	if flow.GetFlowState() != "RUNNING" {
		flow.SetMessageStream(mg.GetNewStream(flow.Id))
		flow.Start()
	}

}

func (mg *Manager) StopFlow(id string) {
	log.Info("Unloading flow , id = ",id)
	if mg.GetFlowById(id) == nil {
		log.Info("Can find flow by id = ",id)
		return
	}
	if mg.GetFlowById(id).GetFlowState() != "RUNNING" {
		log.Info("Flow is not running , nothing to stop.")
		return
	}
	mg.GetFlowById(id).Stop()
	close(mg.msgStreams[id])
	delete(mg.msgStreams,id)
	log.Infof("Flow with Id = %s is unloaded",id)
}

func (mg *Manager) DeleteFlowFromRegistry(id string,cleanRegistry bool) {

	for i :=range mg.flowRegistry {
		if mg.flowRegistry[i].Id == id {
			if cleanRegistry {
				mg.flowRegistry[i].CleanupBeforeDelete()
			}

			mg.flowRegistry = append(mg.flowRegistry[:i], mg.flowRegistry[i+1:]...)
			break
		}
	}
}

func (mg *Manager) DeleteFlowFromStorage(id string) {
	flow := mg.GetFlowById(id)
	if flow == nil {
		return
	}
	mg.StopFlow(id)
	mg.DeleteFlowFromRegistry(id,true)
	os.Remove(mg.GetFlowFileNameById(id))
}


func (mg *Manager) GetConnectorRegistry() *connector.Registry {
	return &mg.connectorRegistry
}