package flow

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow"
	"github.com/thingsplex/tpflow/connector"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/trigger/fimp"
	"github.com/thingsplex/tpflow/utils"
	"runtime/debug"
	"time"

	//thingsplexmodel "github.com/thingsplex/thingsplex/model"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Manager struct {
	flowRegistry      []*Flow
	msgStreams        map[string]model.MsgPipeline
	globalContext     *model.Context
	config            tpflow.Configs
	connectorRegistry connector.Registry
}

type FlowListItem struct {
	Id             string
	Name           string
	Group          string
	Description    string
	State          string
	TriggerCounter int64
	ErrorCounter   int64
	Stats          *model.FlowStatsReport
	IsDisabled     bool
}

func NewManager(config tpflow.Configs) (*Manager, error) {
	var err error
	man := Manager{config: config}
	man.msgStreams = make(map[string]model.MsgPipeline)
	man.flowRegistry = make([]*Flow, 0)
	man.globalContext, err = model.NewContextDB(config.ContextStorageDir)
	if err !=nil {
		log.Error("Can't initialize context DB")
		return nil,err
	}
	man.globalContext.RegisterFlow("global")
	man.connectorRegistry = *connector.NewRegistry(config.ConnectorStorageDir)
	man.connectorRegistry.LoadInstancesFromDisk()
	return &man, err
}

func (mg *Manager) GenerateNewFlow() model.FlowMeta {
	fl := model.FlowMeta{}
	fl.Nodes = []model.MetaNode{{Id: "1", Type: "trigger", Label: "no label", Config: fimp.TriggerConfig{Timeout: 0, ValueFilter: model.Variable{}, IsValueFilterEnabled: false}}}
	fl.Id = utils.GenerateId(15)
	fl.ClassId = fl.Id
	fl.CreatedAt = time.Now()
	fl.UpdatedAt = fl.CreatedAt
	return fl
}

func (mg *Manager) GetNewStream(Id string) model.MsgPipeline {
	msgStream := make(model.MsgPipeline, 10)
	mg.msgStreams[Id] = msgStream
	return msgStream
}

func (mg *Manager) GetGlobalContext() *model.Context {
	return mg.globalContext
}

func (mg *Manager) GetFlowFileNameById(id string) string {
	return filepath.Join(mg.config.FlowStorageDir, id+".json")
}

func (mg *Manager) LoadAllFlowsFromStorage() error {

	dirs := []string{ mg.config.FlowStorageDir,filepath.Join(mg.config.FlowStorageDir,"defaults")}

	for _,dir := range dirs {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Error(err)
			return err
		}
		for _, file := range files {
			if strings.Contains(file.Name(), ".json") {
				err :=  mg.LoadFlowFromFile(filepath.Join(dir, file.Name()))
				if err != nil {
					log.Errorf("Flow %s failed to load from file ",file.Name())
					continue
				}
				flowId := strings.Replace(file.Name(), ".json", "", 1)
				flow := mg.GetFlowById(flowId)
				if flow == nil {
					log.Errorf("Flow %s failed to load",flowId)
					continue
				}
				if !flow.FlowMeta.IsDisabled {
					mg.StartFlow(flowId)
				}
			}
		}
	}
	return nil
}

func (mg *Manager) LoadFlowFromFile(fileName string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("-------ERROR 1 Flow can't be loaded from file : ", r)
			debug.PrintStack()
		}
	}()
	log.Info("<FlMan> Loading flow from file : ", fileName)
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Error("<FlMan> Can't open Flow file.")
		return err
	}
	mg.LoadFlowFromJson(file)
	return nil
}

func (mg *Manager) LoadFlowFromJson(flowJsonDef []byte) error {
	flowMeta := model.FlowMeta{}
	err := json.Unmarshal(flowJsonDef, &flowMeta)
	if err != nil {
		log.Error("<FlMan> Can't unmarshel DB file.")
		return err
	}
	return mg.AddMetaFlowToRegistry(flowMeta)
}

func (mg *Manager) AddMetaFlowToRegistry(flowMeta model.FlowMeta) error {
	flow := NewFlow(flowMeta, mg.globalContext)
	flow.SetStoragePath(mg.config.FlowStorageDir)
	flow.SetExternalLibsDir(mg.config.ExternalLibsDir)
	flow.SetConnectorRegistry(&mg.connectorRegistry)
	mg.flowRegistry = append(mg.flowRegistry, flow)
	return nil
}

func (mg *Manager) UpdateFlowFromBinJson(id string, flowJsonDef []byte) error {
	flowMeta := model.FlowMeta{}
	err := json.Unmarshal(flowJsonDef, &flowMeta)
	if err != nil {
		log.Error("<FlMan> The flow is not updated . Incompatible format")
		return err
	}
	if flowMeta.IsDefault {
		log.Error("<FlMan> Default flows are constant.Operation skipped.")
		return err
	}
	mg.StopFlow(id)
	mg.DeleteFlowFromRegistry(id, false)
	flowMeta.UpdatedAt = time.Now()
	err = mg.AddMetaFlowToRegistry(flowMeta)
	if err != nil {
		log.Error("<FlMan> The flow is not updated . Incompatible format . Err:",err)
		return err
	}
	err = mg.SaveFlowToStorage(id)
	if err != nil {
		log.Error("<FlMan> Can't save the flow to disk . Err:",err)
		return err
	}
	if !flowMeta.IsDisabled {
		mg.StartFlow(id)
	}
	log.Infof("<FlMan> Flow %s is valid , saving to disk",id)
	return nil
}

func (mg *Manager) ReloadFlowFromStorage(id string) error {
	mg.StopFlow(id)
	return mg.LoadFlowFromFile(mg.GetFlowFileNameById(id))
}

func (mg *Manager) ImportFlow(flowJsonDef []byte) error {
	newId := utils.GenerateId(15)
	flowMeta := model.FlowMeta{}
	err := json.Unmarshal(flowJsonDef, &flowMeta)
	if err != nil {
		log.Error("<FlMan> Can't unmarshel imported flow 1.")
		return err
	}
	oldId := flowMeta.Id
	// Replacing all old flow id's with new id.
	flowAsString := string(flowJsonDef)
	flowAsStringUpdated := strings.Replace(flowAsString, oldId, newId, -1)

	// Updating class id
	err = json.Unmarshal([]byte(flowAsStringUpdated), &flowMeta)
	if err != nil {
		log.Error("<FlMan> Can't unmarshel imported flow 2.")
		return err
	}

	flowMeta.ClassId = oldId
	flowMeta.CreatedAt = time.Now()
	flowMeta.UpdatedAt = flowMeta.CreatedAt
	flowMetaByte,err := json.Marshal(&flowMeta)

	if err != nil {
		log.Error("<FlMan> Can't marshar imported flow ")
		return err
	}

	fileName := mg.GetFlowFileNameById(newId)
	err = ioutil.WriteFile(fileName, flowMetaByte, 0644)
	if err != nil {
		log.Error("Can't save flow to file . Error : ", err)
		return err
	}
	log.Debugf("<FlMan> Flow is imported")
	return mg.LoadFlowFromFile(fileName)
}

func (mg *Manager) SaveFlowToStorage(id string) error {
	flow := mg.GetFlowById(id)
	flowMetaByte,err := json.Marshal(flow.FlowMeta)
	if err != nil {
		log.Error("<FlMan> Can't marshar imported flow ")
		return err
	}
	fileName := mg.GetFlowFileNameById(id)
	log.Debugf("<FlMan> Saving flow to file %s , data size %d :", fileName, len(flowMetaByte))
	err = ioutil.WriteFile(fileName, flowMetaByte, 0644)
	if err != nil {
		log.Error("Can't save flow to file . Error : ", err)
		return err
	}
	log.Debugf("<FlMan> Flow saved")
	return nil
}

func (mg *Manager) GetFlowById(id string) *Flow {
	for i := range mg.flowRegistry {
		if mg.flowRegistry[i].Id == id {
			return mg.flowRegistry[i]
		}
	}
	return nil
}

func (mg *Manager) GetFlowBySettings(settings map[string]model.Setting) *Flow {
	for i := range mg.flowRegistry {
		fullMatch := true
		log.Debug("Flow ",mg.flowRegistry[i].Name)
		for si := range settings {
			if mg.flowRegistry[i].FlowMeta != nil {
				flSet := mg.flowRegistry[i].FlowMeta.Settings[si]
				sSet := settings[si]
				log.Debugf( "key = %s, flSet = %s , sSet = %s",si,flSet.String(),sSet.String())
				if flSet.String() != sSet.String() {
					fullMatch = false
				}
			}else {
				fullMatch = false
			}
		}
		if fullMatch {
			return mg.flowRegistry[i]
		}
	}
	return nil
}

func (mg *Manager) GetFlowList() []FlowListItem {
	response := make([]FlowListItem, len(mg.flowRegistry))
	var c int
	for i := range mg.flowRegistry {
		response[c] = FlowListItem{
			Id:             mg.flowRegistry[i].Id,
			Name:           mg.flowRegistry[i].Name,
			Group:          mg.flowRegistry[i].FlowMeta.Group,
			Description:    mg.flowRegistry[i].Description,
			TriggerCounter: mg.flowRegistry[i].TriggerCounter,
			ErrorCounter:   mg.flowRegistry[i].ErrorCounter,
			State:          mg.flowRegistry[i].opContext.State,
			Stats:          mg.flowRegistry[i].GetFlowStats(),
			IsDisabled:     mg.flowRegistry[i].FlowMeta.IsDisabled,
		}

		c++
	}
	return response
}

func (mg *Manager) ControlFlow(cmd string, flowId string) error {
	switch cmd {
	case "START":
		mg.StartFlow(flowId)
	case "STOP":
		mg.StopFlow(flowId)
	}
	return nil
}

func (mg *Manager) StartFlow(flowId string) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("-------ERROR 2 Flow can't be loaded from file : ", r)
			debug.PrintStack()
		}
	}()
	flow := mg.GetFlowById(flowId)
	if flow.FlowMeta.IsDisabled {
		flow.FlowMeta.IsDisabled = false
		mg.SaveFlowToStorage(flowId)
	}

	if flow != nil {
		if flow.GetFlowState() != "RUNNING" {
			//flow.SetMessageStream(mg.GetNewStream(flow.Id))
			flow.Start()
		}
	}else {
		log.Error("No flow with Id = ",flowId)
	}


}

func (mg *Manager) StopFlow(id string) {
	log.Info("Unloading flow , id = ", id)
	flow := mg.GetFlowById(id)
	if flow == nil {
		log.Info("Can find flow by id = ", id)
		return
	}
	if flow.GetFlowState() != "RUNNING" {
		log.Info("Flow is not running , nothing to stop.")
		return
	}
	flow.Stop()
	if flow.FlowMeta.IsDefault {
		log.Infof("Flow with Id = %s was stopped but NOT disabled", id)
		return
	}
	flow.FlowMeta.IsDisabled = true
	mg.SaveFlowToStorage(id)
	log.Infof("Flow with Id = %s was stopped and disabled", id)
}

func (mg *Manager) DeleteFlowFromRegistry(id string, cleanRegistry bool) {

	for i := range mg.flowRegistry {
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
	if flow.FlowMeta.IsDefault {
		log.Errorf("Flow delete operation is skipped. Default flow can't be deleted ")
		return
	}
	mg.StopFlow(id)
	mg.DeleteFlowFromRegistry(id, true)
	os.Remove(mg.GetFlowFileNameById(id))
}

func (mg *Manager) GetConnectorRegistry() *connector.Registry {
	return &mg.connectorRegistry
}

func (mg *Manager) BackupAll() error{
	sourceDir := strings.ReplaceAll(mg.config.FlowStorageDir,"/flow_storage","")
	targetDir := strings.ReplaceAll(mg.config.FlowStorageDir,"var/flow_storage","backups")
	return utils.BackupDirectory(sourceDir,targetDir,"flow")
}

// FactoryReset deletes all nodes and restarts tpflow
func (mg *Manager) FactoryReset() error{
	files, err := ioutil.ReadDir(mg.config.FlowStorageDir)
	if err != nil {
		log.Error(err)
		return err
	}
	for _, file := range files {
		if !file.IsDir() {
			os.Remove(filepath.Join(mg.config.FlowStorageDir,file.Name()))
		}
	}
	mg.globalContext.FactoryReset()
	return nil
}