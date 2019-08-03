package plugins

import (
	"github.com/thingsplex/tpflow/connector/model"
	"github.com/thingsplex/tpflow/connector/plugins/fimpmqtt"
	"github.com/thingsplex/tpflow/connector/plugins/influxdb"
)

var pluginRegistry = map[string]model.Plugin{
	"influxdb": {Constructor: influxdb.NewConnectorInstance, Config: influxdb.ConnectorConfig{}},
	"fimpmqtt": {Constructor: fimpmqtt.NewConnectorInstance, Config: fimpmqtt.ConnectorConfig{}},
}

func GetPlugin(name string) *model.Plugin {
	plugin, ok := pluginRegistry[name]
	if ok {
		return &plugin
	}
	return nil

}

func RegisterPlugin(name string, plugin model.Plugin) {
	pluginRegistry[name] = plugin
}

func GetConfigurationTemplate(name string) model.Instance {
	inst := model.Instance{}
	if p := GetPlugin(name); p != nil {
		inst.Config = p.Config
	}
	return inst

}

func GetPlugins() map[string]model.Plugin {
	return pluginRegistry
}
