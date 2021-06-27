package connector

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/connector/model"
	"github.com/thingsplex/tpflow/connector/plugins"
	"github.com/thingsplex/tpflow/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Adapter instance has to be created from Flow manager .
// Node should refuse to start if Adapter

type Registry struct {
	instances  []*model.Instance
	configsDir string
}

func NewRegistry(configDir string) *Registry {
	reg := Registry{configsDir: configDir}
	return &reg
}

// AddConnection Adds existing connection as connector instance into the registry
func (reg *Registry) AddConnection(id string, name string, connType string, conn model.ConnInterface,config interface{}) {
	inst := model.Instance{ID: id, Name: name, Plugin: connType, Connection: conn,Config: config}
	reg.instances = append(reg.instances, &inst)
}

// AddInstance Adds existing instance to registry
func (reg *Registry) AddInstance(inst *model.Instance) {
	reg.instances = append(reg.instances, inst)
	log.Infof("<ConnRegistry> Instance was added , id = %s , name = %s ", inst.ID, inst.Name)
}

// CreateInstance Creates instance of connector using one of registered plugins
func (reg *Registry) CreateInstance(inst  model.Instance) model.ConnInterface {
	connPlugin := plugins.GetPlugin(inst.Plugin)
	if connPlugin != nil {
		if inst.ID == "" {
			inst.ID = utils.GenerateId(15)
		}
		inst.Connection = connPlugin.Constructor(inst.Name, inst.Config)
		if inst.Connection.SetDefaults() { // checking if defaults mutated config or not
			inst.Config = inst.Connection.GetConfig() // copying mutated config
			reg.SaveConnectorConfigToDisk(&inst)
		}
		inst.Connection.Init()
		reg.AddInstance(&inst)
		return inst.Connection
	}
	return nil
}

// GetInstance Returns pointer to existing instance of connector
func (reg *Registry) GetInstance(id string) *model.Instance {
	for i := range reg.instances {
		if reg.instances[i].ID == id {
			return reg.instances[i]
		}
	}
	return nil
}

// GetAllInstances Returns pointer to existing instance of connector
func (reg *Registry) GetAllInstances() []model.InstanceView {
	var instList []model.InstanceView
	for i := range reg.instances {
		inst := model.InstanceView{ID: reg.instances[i].ID, Name: reg.instances[i].Name, Plugin: reg.instances[i].Plugin, State: reg.instances[i].Connection.GetState(), Config: reg.instances[i].Config}
		instList = append(instList, inst)
	}
	return instList
}

func (reg *Registry) LoadDefaults()error {
	//configFile := filepath.Join(reg.configsDir,"data","config.json")
	//os.Remove(configFile)
	//log.Info("Config file doesn't exist.Loading default config")
	//defaultConfigFile := filepath.Join(cf.WorkDir,"defaults","config.json")
	//return utils.CopyFile(defaultConfigFile,configFile)
	dataDir := filepath.Join(reg.configsDir,"data")
	files, err := ioutil.ReadDir(dataDir)
	if err != nil {
		log.Error(err)
		return err
	}
	for _, file := range files {

	}
}

func (reg *Registry) LoadInstancesFromDisk() error {
	log.Info("<ConnRegistry> Loading connectors from disk ")
	dataDir := filepath.Join(reg.configsDir,"data")
	files, err := ioutil.ReadDir(dataDir)
	if err != nil {
		log.Error(err)
		return err
	}
	for _, file := range files {
		if strings.Contains(file.Name(), ".json") {
			fileName := filepath.Join(dataDir, file.Name())
			log.Info("<ConnRegistry> Loading connector instance from file : ", fileName)
			file, err := ioutil.ReadFile(fileName)
			if err != nil {
				log.Error("<ConnRegistry> Can't open connector config file.")
				continue
			}
			instConf := model.Instance{ConfigFileName: fileName}
			err = json.Unmarshal(file, &instConf)
			if err != nil {
				log.Error("<ConnRegistry> Can't unmarshel connector file.")
				continue
			}
			reg.CreateInstance(instConf)
		}
	}
	return nil
}

func (reg *Registry) UpdateConnectorConfig(ID string ,config interface{}) error {
	log.Info("<ConnRegistry> Updating connector config . ID =  ",ID)
	inst := reg.GetInstance(ID)
	if inst == nil {
		return fmt.Errorf("conn instance not found")
	}
	inst.Config = config
	inst.Connection.LoadConfig(config)
	return reg.SaveConnectorConfigToDisk(inst)
}

func (reg *Registry) SaveConnectorConfigToDisk(inst *model.Instance)error {
	log.Info("<ConnRegistry> Saving connector to disk , filename: ",inst.ConfigFileName)
	log.Warningf("%+v",inst.Config)
	bp , err := json.Marshal(inst)
	if err != nil {
		return err
	}
	ioutil.WriteFile(inst.ConfigFileName,bp,0777)
	return nil
}