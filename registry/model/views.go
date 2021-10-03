package model

import "github.com/thingsplex/tpflow/flow/context"

type ServiceExtendedView struct {
	Service
	LocationAlias   string `json:"location_alias"`
	LocationType    string `json:"location_type"`
	LocationSubType string `json:"location_sub_type"`
}

type LocationExtendedView struct {
	Location
	Devices []DeviceExtendedView
}

type ThingWithLocationView struct {
	Thing
	LocationAlias string `json:"location_alias"`
}

type DeviceWithLocationView struct {
	Device
	LocationAlias string `json:"location_alias"`
}

type ThingExtendedView struct {
	Thing
	LocationAlias string `json:"location_alias"`
}

type DeviceExtendedView struct {
	Device
	Services      []Service               `json:"services"`
	LocationAlias string                  `json:"location_alias"`
	States        []context.ContextRecord `json:"states"`
}

type InterfaceFlatView struct {
	ThingId            ID       `json:"thing_id"`
	ThingAddress       string   `json:"thing_address"`
	ThingTech          string   `json:"thing_tech"`
	ThingAlias         string   `json:"thing_alias"`
	ServiceId          ID       `json:"service_id"`
	ServiceName        string   `json:"service_name"`
	ServiceAlias       string   `json:"service_alias"`
	ServiceAddress     string   `json:"service_address"`
	InterfaceType      string   `json:"intf_type"`
	InterfaceMsgType   string   `json:"intf_msg_type"`
	InterfaceAddress   string   `json:"intf_address"`
	InterfaceValueType string   `json:"intf_val_type"`
	LocationId         ID       `json:"location_id"`
	LocationAlias      string   `json:"location_alias"`
	LocationType       string   `json:"location_type"`
	Groups             []string `json:"groups"`
}
