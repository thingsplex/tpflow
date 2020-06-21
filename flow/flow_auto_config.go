package flow

import (
	"encoding/json"
	"fmt"
	fgutils "github.com/futurehomeno/fimpgo/utils"
	"github.com/labstack/gommon/log"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/utils"
	"io/ioutil"
	"path/filepath"
	"time"
)

type AutoConfig struct {
	flow *model.FlowMeta
	flowStorage string
}

func (ac *AutoConfig) Flow() *model.FlowMeta {
	return ac.flow
}

func NewAutoConfig(flowStorage string) *AutoConfig {
	return &AutoConfig{flowStorage: flowStorage}
}

func (ac *AutoConfig) LoadFlowFromTemplate(name string) error {
	// load flow from template either from default or templates folder
	newId := utils.GenerateId(15)
	flowFilePath := filepath.Join(ac.flowStorage,"defaults",name+".json")
	if !fgutils.FileExists(flowFilePath) {
		flowFilePath = filepath.Join(ac.flowStorage,"templates",name+".json")
		if !fgutils.FileExists(flowFilePath) {
			return fmt.Errorf("template doesn't exist")
		}
	}
	log.Info("<FlMan> Loading flow from file : ", flowFilePath)
	flowBin, err := ioutil.ReadFile(flowFilePath)
	if err != nil {
		log.Error("<FlMan> Can't open Flow file.")
		return err
	}
	ac.flow = &model.FlowMeta{}
	err = json.Unmarshal(flowBin, ac.flow)
	if err != nil {
		log.Error("<FlMan> Can't unmarshal template flow")
	}
	ac.flow.Id = newId
	ac.flow.ClassId = "template."+name
	ac.flow.CreatedAt = time.Now()
	ac.flow.UpdatedAt = ac.flow.CreatedAt
	return err
}

func (ac *AutoConfig) SaveNewFlow() error {
	flowFilePath := filepath.Join(ac.flowStorage,ac.flow.Id+".json")
	flowMetaByte,err := json.Marshal(ac.flow)
	if err != nil {
		log.Error("<FlMan> Can't marshal flow ")
		return err
	}
	err = ioutil.WriteFile(flowFilePath, flowMetaByte, 0644)
	if err != nil {
		log.Error("Can't save flow to file . Error : ", err)
		return err
	}
	log.Debugf("<FlMan> Flow saved")
	return nil
}

func (ac *AutoConfig) DeleteAutoFlow(name string) {

}

func (ac *AutoConfig) GetSettings()map[string]model.Setting {
	return ac.flow.Settings
}

func (ac *AutoConfig) SetSettings(settings map[string]model.Setting) {
	ac.flow.Settings = settings
}

