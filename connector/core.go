package connector



// Adapter instance has to be created from Flow manager .
// Node should refuse to start if Adapter

type Registry struct {
	instances     []Instance
}

func (reg *Registry) AddInstance(name string,connType string,state string,conn interface{}) {
	inst := Instance{Name:name,ConnectorType:connType,State:state,Connection:conn}
	reg.instances = append(reg.instances,inst)
}

func (reg *Registry) GetInstance(name string,connType string) *Instance {
	for i := range reg.instances {
		if reg.instances[i].Name == name && reg.instances[i].ConnectorType == connType {
			return &reg.instances[i]
		}
	}
	return nil
}