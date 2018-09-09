package plugins

import (
	"github.com/alivinco/tpflow/connector"
	"github.com/alivinco/tpflow/connector/plugins/influxdb"
)

// plugin registry

type Constructor func(config interface{}) connector.ConnInterface

var PluginRegistry = map[string]Constructor{
	"influxdb":      influxdb.NewInfluxdbConnectorInstance,

}
