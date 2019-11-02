package fh

import (
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype/primefimp"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/registry"
	"github.com/thingsplex/tpflow/registry/model"
	"strconv"
	"strings"
)

type VinculumIntegration struct {
	registry *registry.ThingRegistryStore
	msgTransport *fimpgo.MqttTransport
	primeFimpApi * primefimp.ApiClient
	notifyCh chan primefimp.Notify
	
}

func (mg *VinculumIntegration) SetMsgTransport(msgTransport *fimpgo.MqttTransport) {
	mg.msgTransport = msgTransport
}

func NewVinculumIntegration(registry *registry.ThingRegistryStore,msgTransport *fimpgo.MqttTransport) *VinculumIntegration {
	integr :=  &VinculumIntegration{registry: registry}
	integr.msgTransport = msgTransport
	//integr.msgTransport = getTestTransport()
	integr.primeFimpApi = primefimp.NewApiClient("tplex-reg-1",integr.msgTransport ,false)

	// Actual test
	integr.notifyCh = make(chan primefimp.Notify,10)
	integr.primeFimpApi.RegisterChannel("reg-vinc-integr",integr.notifyCh)
	integr.primeFimpApi.StartNotifyRouter()
	go integr.notifyListener()
	return integr
}

func (mg *VinculumIntegration) notifyListener() {

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error("<vin-integr> notify router CRASHED with error :", r)
				}
			}()
			for msg := range mg.notifyCh {
				log.Debug("New vinc msg")
				if msg.Cmd != primefimp.CmdEdit {
					continue
				}
				switch msg.Component {
				case primefimp.ComponentDevice:
					//log.Infof("New notify from device %s",msg.GetDevice().Client.Name)
					mg.ProcessVincDeviceUpdate([]primefimp.Device{*msg.GetDevice()})
				case primefimp.ComponentRoom:
					mg.ProcessVincRoomUpdate([]primefimp.Room{*msg.GetRoom()})

				}
			}
		}()
	}
	log.Info("<vin-integr> Notify listener has terminated")

}


func (mg *VinculumIntegration) ProcessVincDeviceUpdate(devices []primefimp.Device) error {
	log.Info("Updating Things ")
	//TODO:Replace by thing in future
	processedDevices := map[string]bool{}
	//var err error
	for i:= range devices {
		adapter := strings.Replace(devices[i].Fimp.Adapter,"zwave-ad","zw",1)
		_,ok := processedDevices[adapter+":"+devices[i].Fimp.Address]
		thing,_ := mg.registry.GetThingByIntegrationId( strconv.FormatInt(int64(devices[i].ID),16))
		if thing == nil {
			log.Debug("No Thing match by IntegrationId")
			// Try to find device using service topic
			// Request inclusion report here or create Thing bases on information from device object
			thing, _ = mg.registry.GetThingByAddress(adapter, devices[i].Fimp.Address)
			if thing == nil {
				log.Debug("No Thing match by address")
				thing = &model.Thing{}
				thing.Address = devices[i].Fimp.Address
				thing.CommTechnology = adapter
				if devices[i].Client.Name != nil {
					thing.Alias = *devices[i].Client.Name
				}


				if !ok {
					// Requesting inclusion report from adapter to gather more extended info
					responseMsg := fimpgo.NewMessage("cmd.thing.get_inclusion_report", devices[i].Fimp.Adapter, "string", thing.Address, nil, nil, nil)
					addr := fimpgo.Address{MsgType: fimpgo.MsgTypeCmd, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: adapter, ResourceAddress: "1"}
					if mg.msgTransport == nil {
						log.Error("MQTT transport is NULL")
					}
					mg.msgTransport.Publish(&addr, responseMsg)
				}

			} else {
				log.Info("<MqRegInt> Thing already in registry . Updating")
				// Thing is already in registry but doesn't have link with VInculum , setting link by configuring IntegrationId
				thing.IntegrationId = strconv.FormatInt(int64(devices[i].ID),16)
			}
		}else {
			// Thing is already in registry and in sync with vinculum , user has updated Name
			if devices[i].Client.Name != nil {
				thing.Alias = *devices[i].Client.Name
			}
		}
		if devices[i].Room != nil {
			loc , _ := mg.registry.GetLocationByIntegrationId(strconv.FormatInt(int64(*devices[i].Room),16))
			if loc != nil {
				log.Debug("Updating location")
				thing.LocationId = loc.ID
			}
		}

		var thingID model.ID
		if !ok {
			thingID ,_ = mg.registry.UpsertThing(thing)
		}else {
			thingID = thing.ID
		}
		if devices[i].Client.Name != nil && devices[i].Room != nil {
			mg.ProcessVincServiceUpdate(*devices[i].Client.Name,*devices[i].Room,thingID,devices[i].Service)
		}



		processedDevices[adapter+":"+devices[i].Fimp.Address] = true
	}
	return nil
}

func (mg *VinculumIntegration) ProcessVincServiceUpdate(devName string,roomId int ,thingID model.ID,services map[string]primefimp.Service) error {
	loc , _ := mg.registry.GetLocationByIntegrationId(strconv.FormatInt(int64(roomId),16))
	for name,serv := range services {
		regService,_ := mg.registry.GetServiceByAddress(name,serv.Addr)

		if regService == nil {
			regService = &model.Service{}
			regService.Address = serv.Addr
			regService.Name = name
			regService.ParentContainerType = model.ThingContainer

		}
		regService.ParentContainerId = thingID
		if loc != nil {
			regService.LocationId = loc.ID
		}

		regService.Alias = devName
		mg.registry.UpsertService(regService)
	}
	return nil
}

func (mg *VinculumIntegration) ProcessVincRoomUpdate(rooms []primefimp.Room) error {
	log.Info("Updating Location ")
	var err error
	for i:= range rooms {
		loc,_ := mg.registry.GetLocationByIntegrationId( strconv.FormatInt(int64(rooms[i].ID),16))
		if rooms[i].Type != nil && rooms[i].Client.Name != nil {
			if loc == nil {
				// Location doesn't exist in registry
				loc = &model.Location{Type: "room",SubType:*rooms[i].Type,Alias:*rooms[i].Client.Name,IntegrationId:strconv.FormatInt(int64(rooms[i].ID),16)}
			} else {
				loc.Alias = *rooms[i].Client.Name
			}
		}



		if loc.Alias == "" && rooms[i].Type!=nil{
			loc.Alias = *rooms[i].Type
		}
		_,err = mg.registry.UpsertLocation(loc)
		if err != nil {
			log.Error("Can't update Location . err :",err)
		}else {
			log.Debug("Location updated")
		}
	}
	return nil

}

func (mg *VinculumIntegration) SyncDevice() error {
	devices,err := mg.primeFimpApi.GetDevices(false)
	if err != nil {
		return err
	}
	return mg.ProcessVincDeviceUpdate(devices)

}

func (mg *VinculumIntegration) SyncRooms() error {
	rooms,err := mg.primeFimpApi.GetRooms(false)
	if err != nil {
		return err
	}
	return mg.ProcessVincRoomUpdate(rooms)
}


func getTestTransport() *fimpgo.MqttTransport {
	mqtt := fimpgo.NewMqttTransport("tcp://cube.local:1883","tp-registry-test-1","","",true,1,1)
	err := mqtt.Start()
	log.Info(" Test transport Connected")
	if err != nil {
		log.Error("Error connecting to broker ",err)
	}
	return mqtt
}