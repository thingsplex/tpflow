package storage

import (
	"errors"
	"fmt"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype/primefimp"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow"
	"github.com/thingsplex/tpflow/registry/model"
)

type VinculumRegistryStore struct {
	vApi         *primefimp.ApiClient
	msgTransport *fimpgo.MqttTransport
	config       *tpflow.Configs
}

func (r *VinculumRegistryStore) GetConfig() interface{} {
	return nil
}

func (r *VinculumRegistryStore) SetDefaults() bool {
	return false
}

func NewVinculumRegistryStore(config *tpflow.Configs) RegistryStorage {
	return &VinculumRegistryStore{config: config}
}

func (r *VinculumRegistryStore) Connect() error {
	clientId := r.config.MqttClientIdPrefix + "vinc_reg_store"
	r.msgTransport = fimpgo.NewMqttTransport(r.config.MqttServerURI, clientId, r.config.MqttUsername, r.config.MqttPassword, true, 1, 1)
	r.msgTransport.SetGlobalTopicPrefix(r.config.MqttTopicGlobalPrefix)
	err := r.msgTransport.Start()
	log.Info("<MqRegInt> Mqtt transport connected")
	if err != nil {
		log.Error("<MqRegInt> Error connecting to broker : ", err)
		return err
	}
	r.vApi = primefimp.NewApiClient("tpflow_reg", r.msgTransport, false)

	if r.config.IsDevMode {
		log.Info("<MqRegInt> DEV-MODE vinculum DB from file")
		err = r.vApi.LoadVincResponseFromFile("testdata/vinfimp/site-response.json")
		if err != nil {
			log.Error("Can't load site data from file . Error:", err)
		}
	} else {
		go func() {
			r.vApi.ReloadSiteToCache(40)
			r.vApi.StartNotifyRouter()
		}()
	}
	return nil
}

func (r *VinculumRegistryStore) GetBackendName() string {
	return "vinculum"
}

func (r *VinculumRegistryStore) Disconnect() {
	r.vApi.Stop()
}

func (VinculumRegistryStore) GetServiceById(Id model.ID) (*model.Service, error) {
	//log.Warn("GetServiceById NOT implemented !!!!")
	return nil, errors.New("not implemented")
}

func (VinculumRegistryStore) GetServiceByFullAddress(address string) (*model.ServiceExtendedView, error) {
	//log.Warn("GetServiceByFullAddress NOT implemented !!!!")
	return nil, errors.New("not implemented")
}

func (r *VinculumRegistryStore) GetServicesByLocationAndName(locId model.ID, name string) ([]model.Service, error) {
	var result []model.Service
	services , err := r.GetAllServices()
	if err != nil {
		return nil, err
	}
	for i := range services {
		if (services[i].Name == name || name =="" )&& services[i].LocationId == locId {
			result = append(result,services[i])
		}
	}
	return result , err
}

func (r *VinculumRegistryStore) GetLocationById(Id model.ID) (*model.Location, error) {
	locations, err := r.GetAllLocations()
	if err != nil {
		return nil, err
	}

	for i := range locations {
		if locations[i].ID == Id {
			return &locations[i], nil
		}
	}
	return nil, nil
}

func (r *VinculumRegistryStore) GetAllThings() ([]model.Thing, error) {
	site, err := r.vApi.GetSite(true)

	if err != nil {
		return nil, err
	}
	vThings := site.Things
	var things []model.Thing
	for i := range vThings {
		thing := model.Thing{}
		thing.ID = model.ID(vThings[i].ID)
		thing.Alias = vThings[i].Name
		thing.Address = vThings[i].Address
		thing.LocationId = model.ID(vThings[i].RoomID)
		thing.ProductHash, _ = vThings[i].Props["product_hash"]
		thing.ProductId, _ = vThings[i].Props["product_id"]
		thing.ProductName, _ = vThings[i].Props["product_name"]
		for di := range vThings[i].Devices {
			firstDev := site.GetDeviceById(vThings[i].Devices[di])
			if firstDev != nil {
				thing.CommTechnology = firstDev.Fimp.Adapter
				_, isBattery := firstDev.Service["battery"]
				if isBattery {
					thing.PowerSource = "battery"
				} else {
					thing.PowerSource = "ac"
				}
				if thing.LocationId == 0 {
					if firstDev.Room == nil {
						continue
					}
					if *firstDev.Room == 0 {
						continue
					}
					thing.LocationId = model.ID(*firstDev.Room)
				}
			}
		}
		things = append(things, thing)
	}
	return things, nil
}

func (r *VinculumRegistryStore) GetDevicesByThingId(locationId model.ID) ([]model.Device, error) {
	log.Warn("GetDevicesByThingId NOT implemented  !!!!")
	return nil, nil
}

func (r *VinculumRegistryStore) ExtendThingsWithLocation(things []model.Thing) []model.ThingWithLocationView {
	response := make([]model.ThingWithLocationView, len(things))
	site, err := r.vApi.GetSite(true)
	if err != nil {
		return nil
	}
	for i := range things {
		response[i].Thing = things[i]
		room := site.GetRoomById(int(things[i].LocationId))
		if room != nil {
			if room.Alias != "" {
				response[i].LocationAlias = room.Alias
			} else {
				if room.Type != nil {
					response[i].LocationAlias = *room.Type
				}
			}
			continue
		}
		area := site.GetAreaById(int(things[i].LocationId))
		if area != nil {
			response[i].LocationAlias = area.Name
		}

	}
	return response
}

func (r *VinculumRegistryStore) GetAllServices() ([]model.Service, error) {
	site, err := r.vApi.GetSite(true)
	if err != nil {
		return nil, err
	}
	vDevs := site.Devices
	var services []model.Service
	for i := range vDevs {
		for k := range vDevs[i].Service {
			svc := model.Service{Name: k}
			svc.Address = vDevs[i].Service[k].Addr
			if vDevs[i].Client.Name != nil {
				svc.Alias = *vDevs[i].Client.Name
			}
			for _, intfName := range vDevs[i].Service[k].Interfaces {
				intf := model.Interface{
					Type:      "", // can be extracted from inclusion report , but currently not being used anywhere
					MsgType:   intfName,
					ValueType: "", // can be extracted from inclusion report or use MsgType -> ValueType mapping
					Version:   "",
				}
				svc.Interfaces = append(svc.Interfaces, intf)
			}
			svc.ParentContainerId = model.ID(vDevs[i].ID)
			svc.ParentContainerType = model.DeviceContainer
			if vDevs[i].Room != nil {
				svc.LocationId = model.ID(*vDevs[i].Room)
			}
			services = append(services, svc)
		}
	}

	return services, nil
}

func (r *VinculumRegistryStore) GetServicesByDeviceId(Id model.ID)([]model.Service,error) {
	return nil, errors.New("not implemented")
}

func (VinculumRegistryStore) GetThingExtendedViewById(Id model.ID) (*model.ThingExtendedView, error) {
	//log.Warn("GetThingExtendedViewById NOT implemented  !!!!")
	return nil, errors.New("not implemented")
}

func (VinculumRegistryStore) GetServiceByAddress(serviceName string, serviceAddress string) (*model.Service, error) {
	//log.Warn("GetServiceByAddress NOT implemented  !!!!")
	return nil, errors.New("not implemented")
}

func (VinculumRegistryStore) GetExtendedServices(serviceNameFilter string, filterWithoutAlias bool, thingIdFilter model.ID, locationIdFilter model.ID) ([]model.ServiceExtendedView, error) {
	var svcs []model.ServiceExtendedView
	return svcs, errors.New("not implemented")
}

func (r *VinculumRegistryStore) GetAllLocations() ([]model.Location, error) {
	site, err := r.vApi.GetSite(true)
	if err != nil {
		return nil, err
	}
	var locations []model.Location

	for i := range site.Areas {
		loc := model.Location{}
		loc.ID = model.ID(site.Areas[i].ID) * (-1)
		loc.Type = "area"
		loc.Alias = site.Areas[i].Name
		loc.SubType = site.Areas[i].Type
		locations = append(locations, loc)
	}

	for i := range site.Rooms {
		loc := model.Location{}
		loc.ID = model.ID(site.Rooms[i].ID)
		loc.Type = "room"
		loc.Alias = site.Rooms[i].Alias
		if site.Rooms[i].Type != nil {
			loc.SubType = *site.Rooms[i].Type
		}
		if site.Rooms[i].Area != nil {
			loc.ParentID = model.ID(*site.Rooms[i].Area)
		}
		locations = append(locations, loc)
	}
	return locations, nil
}

func (VinculumRegistryStore) GetThingById(Id model.ID) (*model.Thing, error) {
	//log.Warn("GetThingById NOT implemented  !!!!")
	return nil, errors.New("not implemented")
}

func (VinculumRegistryStore) GetThingByAddress(technology string, address string) (*model.Thing, error) {
	//log.Warn("GetThingByAddress NOT implemented  !!!!")
	return nil, errors.New("not implemented")
}

func (VinculumRegistryStore) GetThingExtendedViewByAddress(technology string, address string) (*model.ThingExtendedView, error) {
	//log.Warn("GetThingExtendedViewByAddress NOT implemented  !!!!")
	return nil, errors.New("not implemented")
}

func (r *VinculumRegistryStore) GetThingsByLocationId(locationId model.ID) ([]model.Thing, error) {
	site, err := r.vApi.GetSite(true)
	if err != nil {
		return nil, err
	}
	var things []model.Thing
	//err := st.db.Select(q.Eq("LocationId", locationId)).Find(&things)
	//return things, err
	for i := range site.Things {
		if site.Things[i].RoomID == int(locationId) {
			thing, err := r.GetThingById(model.ID(site.Things[i].RoomID))
			if err != nil {
				things = append(things, *thing)
			}
		}
	}
	return things, nil
}

func (VinculumRegistryStore) GetThingByIntegrationId(id string) (*model.Thing, error) {
	//log.Warn("GetThingByIntegrationId NOT implemented  !!!!")
	return nil, errors.New("not implemented")
}

func (r *VinculumRegistryStore) GetAllDevices() ([]model.Device, error) {
	site, err := r.vApi.GetSite(true)
	if err != nil {
		return nil, err
	}
	vDevices := site.Devices
	var devices []model.Device

	for i := range vDevices {
		device := model.Device{}
		device.ID = model.ID(vDevices[i].ID)
		if vDevices[i].Client.Name != nil {
			device.Alias = *vDevices[i].Client.Name
		}
		if vDevices[i].Room != nil {
			device.LocationId = model.ID(*vDevices[i].Room)
		}

		if vDevices[i].ThingID != nil {
			device.ThingId = model.ID(*vDevices[i].ThingID)
		}
		var dtype, dsubtype string
		if vDevices[i].Type["type"] != nil {
			dtype, _ = vDevices[i].Type["type"].(string)
		}
		if vDevices[i].Type["subtype"] != nil {
			dsubtype, _ = vDevices[i].Type["subtype"].(string)
			dsubtype = "." + dsubtype
		}
		device.Type = fmt.Sprintf("%s%s", dtype, dsubtype)

		devices = append(devices, device)
	}
	return devices, nil
}

func (r *VinculumRegistryStore) GetExtendedDevices() ([]model.DeviceExtendedView, error) {

	site, err := r.vApi.GetSite(true)
	if err != nil {
		return nil, err
	}
	vDevices := site.Devices
	var devices []model.DeviceExtendedView

	for i := range vDevices {
		device := model.DeviceExtendedView{}
		device.ID = model.ID(vDevices[i].ID)
		if vDevices[i].Client.Name != nil {
			device.Alias = *vDevices[i].Client.Name
		}
		if vDevices[i].Room != nil {
			device.LocationId = model.ID(*vDevices[i].Room)
			location, err := r.GetLocationById(device.LocationId)
				if err == nil && location != nil {
					device.LocationAlias = fmt.Sprintf("%s(%s %s)", location.Alias, location.Type, location.SubType)
				}
		}

		if vDevices[i].ThingID != nil {
			device.ThingId = model.ID(*vDevices[i].ThingID)
		}
		var dtype, dsubtype string
		if vDevices[i].Type["type"] != nil {
			dtype, _ = vDevices[i].Type["type"].(string)
		}
		if vDevices[i].Type["subtype"] != nil {
			dsubtype, _ = vDevices[i].Type["subtype"].(string)
			dsubtype = "." + dsubtype
		}
		device.Type = fmt.Sprintf("%s%s", dtype, dsubtype)

		device.Services = make([]model.Service,len(vDevices[i].Service))
		var si int
		for k,v := range vDevices[i].Service {
			device.Services[si].Address = v.Addr
			device.Services[si].Props = v.Props
			device.Services[si].Name = k
			device.Services[si].Enabled = v.Enabled

			for _, intfName := range v.Interfaces {
				intf := model.Interface{MsgType:intfName}
				device.Services[si].Interfaces = append(device.Services[si].Interfaces, intf)
			}
			si++
		}
		devices = append(devices, device)
	}

	return devices, nil

}

func (VinculumRegistryStore) GetDeviceById(Id model.ID) (*model.DeviceExtendedView, error) {
	//log.Warn("GetThingByIntegrationId NOT implemented  !!!!")
	return nil, errors.New("not implemented")
}

func (VinculumRegistryStore) GetDevicesByLocationId(locationId model.ID) ([]model.Device, error) {
	//log.Warn("GetThingByIntegrationId NOT implemented  !!!!")
	return nil, errors.New("not implemented")
}

func (VinculumRegistryStore) GetDeviceByIntegrationId(id string) (*model.Device, error) {
	//log.Warn("GetThingByIntegrationId NOT implemented  !!!!")
	return nil, errors.New("not implemented")
}

func (VinculumRegistryStore) GetLocationByIntegrationId(id string) (*model.Location, error) {
	//log.Warn("GetThingByIntegrationId NOT implemented  !!!!")
	return nil, errors.New("not implemented")
}

func (VinculumRegistryStore) UpsertThing(thing *model.Thing) (model.ID, error) {
	log.Warn("UpsertThing NOT implemented  !!!!")
	return 0, errors.New("not implemented")
}

func (VinculumRegistryStore) UpsertService(service *model.Service) (model.ID, error) {
	log.Warn("UpsertService NOT implemented  !!!!")
	return 0, errors.New("not implemented")
}

func (VinculumRegistryStore) UpsertLocation(location *model.Location) (model.ID, error) {
	log.Warn("UpsertLocation NOT implemented  !!!!")
	return 0, errors.New("not implemented")
}

func (VinculumRegistryStore) DeleteThing(id model.ID) error {
	log.Warn("DeleteThing NOT implemented  !!!!")
	return errors.New("not implemented")
}

func (VinculumRegistryStore) DeleteService(id model.ID) error {
	log.Warn("DeleteService NOT implemented  !!!!")
	return errors.New("not implemented")
}

func (VinculumRegistryStore) DeleteLocation(id model.ID) error {
	log.Warn("DeleteLocation NOT implemented  !!!!")
	return errors.New("not implemented")
}

func (VinculumRegistryStore) ReindexAll() error {
	log.Warn("ReindexAll NOT implemented  !!!!")
	return errors.New("not implemented")
}

func (r *VinculumRegistryStore) Sync() error {
	return r.vApi.ReloadSiteToCache(3)
}

func (VinculumRegistryStore) ClearAll() error {
	log.Warn("ClearAll NOT implemented  !!!!")
	return errors.New("not implemented")
}

func (r *VinculumRegistryStore) GetConnection() interface{} {
	log.Warn("GetConnection NOT implemented  !!!!")
	return errors.New("not implemented")
}

func (VinculumRegistryStore) GetState() string {
	return "RUNNING"
}

func (r *VinculumRegistryStore) LoadConfig(config interface{}) error {
	return nil
}

func (r *VinculumRegistryStore) Init() error {
	return nil
}

func (r *VinculumRegistryStore) Stop() {
	return
}
