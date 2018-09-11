package connector

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/alivinco/tpflow/connector/model"
	"github.com/alivinco/tpflow/connector/plugins"
	"github.com/alivinco/tpflow/utils"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Adapter instance has to be created from Flow manager .
// Node should refuse to start if Adapter

type Registry struct {
	instances     []*model.Instance
	configsDir   string
}

func NewRegistry(configDir string) *Registry {
	reg := Registry{configsDir:configDir}
	return &reg
}

// Adds existing connection as connector instance into the registry
func (reg *Registry) AddConnection(id string ,name string,connType string,state string,conn interface{}) {
	inst := model.Instance{ID:id, Name:name,ConnectorType:connType,State:state,Connection:conn}
	reg.instances = append(reg.instances,&inst)
}

// Adds existing instance to registry
func (reg *Registry) AddInstance(inst *model.Instance) {
	reg.instances = append(reg.instances,inst)
	log.Info("<ConnRegistry> Instance was added , id = %s , name = %s ",inst.ID,inst.Name)
}



// Creates instance of connector using one of registered plugins
func (reg *Registry) CreateInstance(name string , connType string, config interface{}) model.ConnInterface {
 	connPlugin,err := plugins.GetPlugin(connType)
	if err != nil {
		connInstance := connPlugin.Constructor(name ,config)
		reg.AddConnection(utils.GenerateId(15),name,connType,"RUNNING",connInstance)
		return connInstance
	}
	return nil
}

// Returns pointer to existing instance of connector
func (reg *Registry) GetInstance(name string,connType string) *model.Instance {
	for i := range reg.instances {
		if reg.instances[i].Name == name && reg.instances[i].ConnectorType == connType {
			return reg.instances[i]
		}
	}
	return nil
}



func (reg *Registry) LoadInstancesFromDisk() error {
	log.Info("<ConnRegistry> Loading connectors from disk ")

	files, err := ioutil.ReadDir(reg.configsDir)
	if err != nil {
		log.Error(err)
		return err
	}
	for _, file := range files {
		if strings.Contains(file.Name(),".json"){
			fileName := filepath.Join(reg.configsDir,file.Name())
			log.Info("<ConnRegistry> Loading connector instance from file : ",fileName)
			file, err := ioutil.ReadFile(fileName)
			if err != nil {
				log.Error("<ConnRegistry> Can't open connector config file.")
				continue
			}
			inst := model.Instance{}
			err = json.Unmarshal(file, &inst)
			if err != nil {
				log.Error("<ConnRegistry> Can't unmarshel connector file.")
				continue
			}
			reg.AddInstance(&inst)

		}
	}
	return nil
}