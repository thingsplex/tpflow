package model

import "time"

type ID int

const IDnil = 0

const (
	AppContainer          = "app"
	ThingContainer        = "thing"
	DeviceContainer       = "dev"
	AttributeUpdatedByCmd = 1
	AttributeUpdatedByEvt = 2
)

type ThingRegistry struct {
	Things    []Thing
	Locations []Location
}

type AttributeValueContainer struct {
	Value     interface{}
	ValueType string
	UpdatedAt time.Time
	UpdatedBy int
}

type Thing struct {
	ID                ID                                `json:"id" storm:"id,increment,index"`
	IntegrationId     string                            `json:"integr_id" storm:"index"`
	Address           string                            `json:"address" storm:"index"`
	Type              string                            `json:"type"`
	ProductHash       string                            `json:"product_hash"`
	Alias             string                            `json:"alias"`
	CommTechnology    string                            `json:"comm_tech" storm:"index"`
	ProductId         string                            `json:"product_id"`
	ProductName       string                            `json:"product_name"`
	ManufacturerId    string                            `json:"manufacturer_id"`
	DeviceId          string                            `json:"device_id"`
	HwVersion         string                            `json:"hw_ver"`
	SwVersion         string                            `json:"sw_ver"`
	PowerSource       string                            `json:"power_source"`
	WakeUpInterval    string                            `json:"wakeup_interval"`
	Security          string                            `json:"security"`
	Tags              []string                          `json:"tags"`
	LocationId        ID                                `json:"location_id" storm:"index"`
	PropSets          map[string]map[string]interface{} `json:"prop_set"`
	TechSpecificProps map[string]string                 `json:"tech_specific_props"`
	UpdatedAt         time.Time                         `json:"updated_at"`
}

/*
 "id": "1",
      "locationRef": null,
      "metadata": null,
      "model": "zw_398_3_13",
      "modelAlias": "",
      "name": "Experimental smoke sensor",
      "origin": "00000000e7fd6dd3",
      "services": [
        {
          "address": "/rt:dev/rn:zw/ad:1/sv:basic/ad:9_0",
          "enabled": true,
          "interfaces": [
            "cmd.lvl.get_report",
            "cmd.lvl.set",
            "evt.lvl.report"
          ],
          "metadata": null,
          "name": "basic",
          "props": {
            "is_secure": false,
            "is_unsecure": true
          }
        },
        {
          "address": "/rt:dev/rn:zw/ad:1/sv:battery/ad:9_0",
          "enabled": true,
          "interfaces": [
            "cmd.lvl.get_report",
            "evt.alarm.report",
            "evt.lvl.report"
          ],
          "metadata": null,
          "name": "battery",
          "props": {
            "is_secure": false,
            "is_unsecure": true
          }
        },
        {
          "address": "/rt:dev/rn:zw/ad:1/sv:dev_sys/ad:9_0",
          "enabled": true,
          "interfaces": [
            "cmd.group.add_members",
            "cmd.group.delete_members",
            "cmd.group.get_members",
            "cmd.ping.send",
            "evt.group.members_report",
            "evt.ping.report"
          ],
          "metadata": null,
          "name": "dev_sys",
          "props": {
            "is_secure": false,
            "is_unsecure": true
          }
        }
      ],
      "thingRef": {
        "id": "1"
      },
      "type": {
        "subtype": null,
        "supported": {
          "battery": []
        },
        "type": "battery"
      }
*/

type Device struct {
	ID       ID     `json:"id" storm:"id,increment"`
	ThingId  ID     `json:"thing_id" `
	LocationId ID   `json:"location_id"`
	Alias    string `json:"alias"`
	Type     string `json:"type"`
}

type App struct {
	ID       ID        `json:"id" storm:"id,increment"`
	Services []Service `json:"services"`
}

type Adapter struct {
	ID       ID        `json:"id" storm:"id,increment"`
	Services []Service `json:"services"`
}

type Bridge struct {
	ID       ID        `json:"id" storm:"id,increment"`
	Services []Service `json:"services"`
}

type Service struct {
	ID                  ID                                 `json:"id,omitempty"  storm:"id,increment,index"`
	IntegrationId       string                             `json:"integr_id,omitempty" storm:"index"`
	ParentContainerId   ID                                 `json:"container_id,omitempty" storm:"index"`
	ParentContainerType string                             `json:"container_type,omitempty" storm:"index"`
	Name                string                             `json:"name" storm:"index"`
	Enabled             bool                               `json:"enabled"`
	Alias               string                             `json:"alias,omitempty"`
	Address             string                             `json:"address" storm:"index"`
	Groups              []string                           `json:"groups,omitempty"`
	LocationId          ID                                 `json:"location_id,omitempty" storm:"index"`
	Props               map[string]interface{}             `json:"props,omitempty"`
	Tags                []string                           `json:"tags,omitempty"`
	Interfaces          []Interface                        `json:"interfaces"`
	Attributes          map[string]AttributeValueContainer `json:"attributes,omitempty"`
}

type Interface struct {
	Type      string `json:"intf_t,omitempty"`
	MsgType   string `json:"msg_t"`
	ValueType string `json:"val_t,omitempty"`
	Version   string `json:"ver,omitempty"`
}

type Location struct {
	ID             ID         `json:"id" storm:"id,increment,index"`
	IntegrationId  string     `json:"integr_id"`
	Type           string     `json:"type"`
	SubType        string     `json:"sub_type"`
	Alias          string     `json:"alias"`
	Address        string     `json:"address"`
	Longitude      float64    `json:"long"`
	Latitude       float64    `json:"lat"`
	Image          string     `json:"image"`
	ChildLocations []Location `json:"child_locations"`
	ParentID       ID         `json:"parent_id"`
	State          string     `json:"state"`
}


// for vinculum rooms we'll use positive IDs , for vinculum Areas we'll use negative numbers
// id > 0 -> Room , id < 0 -> Area
