package plugins

import (
	"github.com/thingsplex/tpflow/connector/model"
	"github.com/thingsplex/tpflow/connector/plugins/fimpmqtt"
	"github.com/thingsplex/tpflow/connector/plugins/http"
	"github.com/thingsplex/tpflow/connector/plugins/timeseries"
	"github.com/thingsplex/tpflow/connector/plugins/websocket"
)

var pluginRegistry = map[string]model.Plugin{
	"timeseries":       {Constructor: timeseries.NewConnectorInstance, Config: fimpmqtt.ConnectorConfig{}},
	"fimpmqtt":         {Constructor: fimpmqtt.NewConnectorInstance, Config: fimpmqtt.ConnectorConfig{}},
	"httpserver":       {Constructor: http.NewConnectorInstance, Config: http.ConnectorConfig{}},
	"websocket-client": {Constructor: websocket.NewConnectorInstance, Config: websocket.ConnectorConfig{}},
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
