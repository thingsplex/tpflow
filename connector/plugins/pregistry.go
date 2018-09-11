package plugins

import (
	"errors"
	"github.com/alivinco/tpflow/connector/model"
	"github.com/alivinco/tpflow/connector/plugins/influxdb"
)




var pluginRegistry = map[string]model.Plugin{
	"influxdb": {Constructor: influxdb.NewConnectorInstance, Config: influxdb.ConnectorConfig{}},

}

func GetPlugin(name string) (model.Plugin,error) {
	plugin , ok := pluginRegistry[name]
	if ok {
		return plugin,nil
	}
	return model.Plugin{},errors.New("Plugin not found")

}

func RegisterPlugin(name string ,plugin model.Plugin) {
	pluginRegistry[name] = plugin
}