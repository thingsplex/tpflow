package flow

import (
	"encoding/json"
	"github.com/labstack/gommon/log"
	"github.com/thingsplex/tpflow/model"
	"io/ioutil"
	"path/filepath"
)

type AutoConfig struct {
	flow *model.FlowMeta
	flowStorage string
}

func NewAutoConfig(flowStorage string) *AutoConfig {
	return &AutoConfig{flowStorage: flowStorage}
}

func (ac *AutoConfig) LoadFlowFromTemplate(name string) error {
	// load flow from template either from default or templates folder
	flowFilePath := filepath.Join(ac.flowStorage,"defaults",name+".json")
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
	return err
}

func (ac *AutoConfig) SaveNewFlow(name string) error {
	flowFilePath := filepath.Join(ac.flowStorage,name+".json")
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

func (ac *AutoConfig) GetSettings()map[string]string {
	return ac.flow.Settings
}

func (ac *AutoConfig) SetSettings(settings map[string]string) {
	ac.flow.Settings = settings
}

