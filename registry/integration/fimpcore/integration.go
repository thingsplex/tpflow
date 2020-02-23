package fimpcore

import (
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow"
	"github.com/thingsplex/tpflow/registry/integration/fh"
	"github.com/thingsplex/tpflow/registry/model"
	"github.com/thingsplex/tpflow/registry/storage"
	"runtime/debug"
	"strconv"
)

type MqttIntegration struct {
	msgTransport *fimpgo.MqttTransport
	config       *tpflow.Configs
	registry     storage.RegistryStorage
	vincIntegr   *fh.VinculumIntegration
}

func NewMqttIntegration(config *tpflow.Configs, registry storage.RegistryStorage) *MqttIntegration {
	int := MqttIntegration{config: config, registry: registry}
	return &int
}
func (mg *MqttIntegration) InitMessagingTransport() {
	clientId := mg.config.MqttClientIdPrefix + "things_registry"
	mg.msgTransport = fimpgo.NewMqttTransport(mg.config.MqttServerURI, clientId, mg.config.MqttUsername, mg.config.MqttPassword, true, 1, 1)
	mg.msgTransport.SetGlobalTopicPrefix(mg.config.MqttTopicGlobalPrefix)
	err := mg.msgTransport.Start()
	log.Info("<MqRegInt> Mqtt transport connected")
	if err != nil {
		log.Error("<MqRegInt> Error connecting to broker : ", err)
	}
	mg.msgTransport.SetMessageHandler(mg.onMqttMessage)
	mg.msgTransport.Subscribe("pt:j1/mt:evt/rt:app/+/+")
	mg.msgTransport.Subscribe("pt:j1/mt:evt/rt:ad/+/+")
	mg.msgTransport.Subscribe("pt:j1/mt:cmd/rt:app/rn:registry/ad:1")
	mg.msgTransport.Subscribe("pt:j1/mt:evt/rt:app/rn:vinculum/ad:1")
	mg.vincIntegr = fh.NewVinculumIntegration(mg.registry,mg.msgTransport)

}

func (mg *MqttIntegration) RequestInclusionReport(adapter string, addr string) {
	reqMsg := fimpgo.NewStringMessage("cmd.thing.get_inclusion_report", adapter, addr, nil, nil, nil)
	if adapter == "zwave-ad" {
		adapter = "zw"
	}
	tAddr := fimpgo.Address{MsgType: fimpgo.MsgTypeCmd, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: adapter, ResourceAddress: "1"}
	mg.msgTransport.Publish(&tAddr, reqMsg)
}

func (mg *MqttIntegration) onMqttMessage(topic string, addr *fimpgo.Address, iotMsg *fimpgo.FimpMessage, rawMessage []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("<MqRegInt> MqttIntegration process CRASHED with error : ", r)
			log.Errorf("<MqRegInt> Crashed while processing message from topic = %s msgType = %s", r, addr.MsgType)
			debug.PrintStack()
		}
	}()

	switch iotMsg.Type {
	case "evt.thing.inclusion_report":
		mg.processInclusionReport(iotMsg)
	case "evt.thing.exclusion_report":
		tech := addr.ResourceName
		mg.processExclusionReport(iotMsg, tech)
	case "cmd.registry.sync_devices":
		go mg.vincIntegr.SyncDevice()
	case "cmd.registry.sync_rooms":
		go mg.vincIntegr.SyncRooms()
	case "cmd.service.get_list":
		//  pt:j1/mt:cmd/rt:app/rn:registry/ad:1
		//  {"serv":"registry","type":"cmd.service.get_list","val_t":"str_map","val":{"serviceName":"out_bin_switch","filterWithoutAlias":"true"},"props":null,"tags":null,"uid":"1234455"}
		filters, err := iotMsg.GetStrMapValue()
		if err == nil {
			var filterWithoutAlias bool
			var locationId, thingId int

			locationIdStr, _ := filters["locationId"]
			thingIdStr, _ := filters["thingId"]
			locationId, _ = strconv.Atoi(locationIdStr)
			thingId, _ = strconv.Atoi(thingIdStr)

			serviceName, _ := filters["serviceName"]
			filterFithoutAliasStr, filterOk := filters["filterWithoutAlias"]
			if filterOk {
				if filterFithoutAliasStr == "true" {
					filterWithoutAlias = true
				}
			}
			//if ok {
			response, err := mg.registry.GetExtendedServices(serviceName, filterWithoutAlias, model.ID(thingId), model.ID(locationId))
			if err != nil {
				log.Error("<MqRegInt> Can get services .Err :", err)
			}
			responseMsg := fimpgo.NewMessage("evt.service.list", "registry", "object", response, nil, nil, iotMsg)
			addr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeApp, ResourceName: "registry", ResourceAddress: "1"}
			mg.msgTransport.Publish(&addr, responseMsg)
			//}
		} else {
			log.Error("<MqRegInt> Can't parse value. Error :", err)
		}
	default:
		//log.Info("<MqRegInt> Unsupported message type :", iotMsg.Type)
	}
}

func (mg *MqttIntegration) processInclusionReport(msg *fimpgo.FimpMessage) error {
	log.Info("<MqRegInt> New inclusion report")
	inclReport := fimptype.ThingInclusionReport{}
	err := msg.GetObjectValue(&inclReport)
	log.Debugf("%+v\n", err)
	log.Debugf("%+v\n", inclReport)
	var isUpdateOp bool
	if inclReport.CommTechnology != "" && inclReport.Address != "" {
		var thing *model.Thing
		thing, err := mg.registry.GetThingByAddress(inclReport.CommTechnology, inclReport.Address)
		if err != nil {
			thing = &model.Thing{}
			thing.Alias = inclReport.Alias
		} else {
			isUpdateOp = true
			log.Info("<MqRegInt> Thing already in registry . Updating")
		}
		thing.Address = inclReport.Address
		thing.CommTechnology = inclReport.CommTechnology
		thing.DeviceId = inclReport.DeviceId
		thing.HwVersion = inclReport.HwVersion
		thing.SwVersion = inclReport.SwVersion
		thing.ManufacturerId = inclReport.ManufacturerId
		thing.ProductId = inclReport.ProductId
		thing.ProductHash = inclReport.ProductHash
		thing.ProductName = inclReport.ProductName
		thing.PowerSource = inclReport.PowerSource
		thing.Tags = inclReport.Tags
		thing.PropSets = inclReport.PropSets
		thing.TechSpecificProps = inclReport.TechSpecificProps
		thing.WakeUpInterval = inclReport.WakeUpInterval
		thing.Security = inclReport.Security
		thingId, err := mg.registry.UpsertThing(thing)
		if err != nil {
			log.Error("<MqRegInt> Can't insert new Thing . Error: ", err)
		}
		for i := range inclReport.Services {
			service := model.Service{}
			service.Name = inclReport.Services[i].Name
			service.Address = inclReport.Services[i].Address
			service.Enabled = inclReport.Services[i].Enabled
			if isUpdateOp {
				service.Alias = inclReport.Services[i].Alias
			}
			service.Tags = inclReport.Services[i].Tags
			service.Props = inclReport.Services[i].Props
			service.Groups = inclReport.Services[i].Groups
			service.Interfaces = make([]model.Interface, len(inclReport.Services[i].Interfaces))
			service.ParentContainerId = thingId
			service.ParentContainerType = model.ThingContainer
			for iIntf := range inclReport.Services[i].Interfaces {
				service.Interfaces[iIntf] = model.Interface(inclReport.Services[i].Interfaces[iIntf])
			}
			mg.registry.UpsertService(&service)
		}

	} else {
		log.Error("<MqRegInt> Either address or commTech is empty ")
	}
	return nil

}

func (mg *MqttIntegration) processExclusionReport(msg *fimpgo.FimpMessage, technology string) error {
	log.Info("<MqRegInt> New inclusion report from technology = ", technology)
	exThing := model.Thing{}
	err := msg.GetObjectValue(&exThing)
	if err != nil {
		log.Info("<MqRegInt> Exclusion report can't be processed . Error : ", err)
		return err
	}
	thing, err := mg.registry.GetThingByAddress(technology, exThing.Address)
	if err != nil {
		return errors.New("Can't find the thing to be deleted")
	}
	mg.registry.DeleteThing(thing.ID)
	log.Infof("Thing with address = %s , tech = %s was deleted.", exThing.Address, technology)
	return nil
}

