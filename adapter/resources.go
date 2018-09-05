package adapter

import "github.com/alivinco/tpflow/registry"
import  influx "github.com/influxdata/influxdb/client/v2"

type Adapters struct {
	Registry *registry.ThingRegistryStore
	InfluxDb map[string]influx.Client

}
