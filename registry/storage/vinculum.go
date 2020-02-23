package storage

import (
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype/primefimp"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow"
	"github.com/thingsplex/tpflow/registry/model"
)

type VinculumRegistryStore struct{
	vApi *primefimp.ApiClient
	msgTransport *fimpgo.MqttTransport
	config       *tpflow.Configs
}

func NewVinculumRegistryStore(config *tpflow.Configs) RegistryStorage {
	return &VinculumRegistryStore{config:config}
}

func (r *VinculumRegistryStore) Connect() error {
	clientId := r.config.MqttClientIdPrefix + "things_registry"
	r.msgTransport = fimpgo.NewMqttTransport(r.config.MqttServerURI, clientId, r.config.MqttUsername, r.config.MqttPassword, true, 1, 1)
	r.msgTransport.SetGlobalTopicPrefix(r.config.MqttTopicGlobalPrefix)
	err := r.msgTransport.Start()
	log.Info("<MqRegInt> Mqtt transport connected")
	if err != nil {
		log.Error("<MqRegInt> Error connecting to broker : ", err)
		return err
	}
	r.vApi = primefimp.NewApiClient("tpflow_reg",r.msgTransport,false)
	err = r.vApi.LoadVincResponseFromFile("testdata/vinfimp/site-response.json")
	if err != nil {
		log.Error("Can't load site data from file . Error:",err)
	}
	//r.vApi.StartNotifyRouter()
	return nil
}

func (r *VinculumRegistryStore) Disconnect() {
	r.vApi.Stop()
}

func (VinculumRegistryStore) GetServiceById(Id model.ID) (*model.Service, error) {
	panic("implement me")
}

func (VinculumRegistryStore) GetServiceByFullAddress(address string) (*model.ServiceExtendedView, error) {
	panic("implement me")
}

func (VinculumRegistryStore) GetLocationById(Id model.ID) (*model.Location, error) {
	panic("implement me")
}

func (r *VinculumRegistryStore) GetAllThings() ([]model.Thing, error) {
	site , err := r.vApi.GetSite(true)
	if err != nil {
		return nil,err
	}
	vThings := site.Things
	var things []model.Thing

	for i := range vThings {
		thing := model.Thing{}
		thing.ID = model.ID(vThings[i].ID)
		thing.Alias = vThings[i].Name
		thing.Address = vThings[i].Address
		thing.LocationId =  model.ID(vThings[i].RoomID)
		thing.ProductHash,_ = vThings[i].Props["product_hash"]
		thing.ProductId,_ = vThings[i].Props["product_id"]
		thing.ProductName,_ = vThings[i].Props["product_name"]
		things = append(things,thing)
	}
	return things,nil
}

func (r *VinculumRegistryStore) GetDevicesByThingId(locationId model.ID) ([]model.Device, error) {
	panic("implement me")
}

func (r *VinculumRegistryStore) ExtendThingsWithLocation(things []model.Thing) []model.ThingWithLocationView {
	response := make([]model.ThingWithLocationView, len(things))
	site , err := r.vApi.GetSite(true)
	if err != nil {
		return nil
	}
	for i := range things {
		response[i].Thing = things[i]
		room := site.GetRoomById(int(things[i].LocationId))
		if room != nil {
			if room.Alias != "" {
				response[i].LocationAlias = room.Alias
			}else {
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
	site,err := r.vApi.GetSite(true)
	if err != nil {
		return nil,err
	}
	vDevs := site.Devices
	var services []model.Service
	for i := range vDevs {
		for k := range vDevs[i].Service {
			svc := model.Service{Name:k}
			svc.Address =  vDevs[i].Service[k].Addr
			if vDevs[i].Client.Name != nil {
				svc.Alias = *vDevs[i].Client.Name
			}
			for _,intfName := range vDevs[i].Service[k].Interfaces {
				intf := model.Interface{
					Type:      "", // can be extracted from inclusion report , but currently not being used anywhere
					MsgType:   intfName,
					ValueType: "", // can be extracted from inclusion report or use MsgType -> ValueType mapping
					Version:   "",
				}
				svc.Interfaces = append(svc.Interfaces,intf)
			}
			svc.ParentContainerId = model.ID(vDevs[i].ID)
			svc.ParentContainerType = model.DeviceContainer
			if vDevs[i].Room != nil {
				svc.LocationId = model.ID(*vDevs[i].Room)
			}
			services = append(services,svc)
		}
	}
	return services,nil
}

func (VinculumRegistryStore) GetThingExtendedViewById(Id model.ID) (*model.ThingExtendedView, error) {
	panic("implement me")
}

func (VinculumRegistryStore) GetServiceByAddress(serviceName string, serviceAddress string) (*model.Service, error) {
	panic("implement me")
}

func (VinculumRegistryStore) GetExtendedServices(serviceNameFilter string, filterWithoutAlias bool, thingIdFilter model.ID, locationIdFilter model.ID) ([]model.ServiceExtendedView, error) {
	var svcs []model.ServiceExtendedView
	return svcs,nil
}

func (r *VinculumRegistryStore) GetAllLocations() ([]model.Location, error) {
	site,err := r.vApi.GetSite(true)
	if err != nil {
		return nil,err
	}
	var locations []model.Location

	for i := range site.Areas {
		loc := model.Location{}
		loc.ID = model.ID(site.Areas[i].ID)*(-1)
		loc.Type = "area"
		loc.Alias = site.Areas[i].Name
		loc.SubType = site.Areas[i].Type
		locations = append(locations,loc)
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
		locations = append(locations,loc)
	}
	return locations,nil
}

func (VinculumRegistryStore) GetThingById(Id model.ID) (*model.Thing, error) {
	panic("implement me")
}

func (VinculumRegistryStore) GetThingByAddress(technology string, address string) (*model.Thing, error) {
	panic("implement me")
}

func (VinculumRegistryStore) GetThingExtendedViewByAddress(technology string, address string) (*model.ThingExtendedView, error) {
	panic("implement me")
}

func (r *VinculumRegistryStore) GetThingsByLocationId(locationId model.ID) ([]model.Thing, error) {
	site,err := r.vApi.GetSite(true)
	if err != nil {
		return nil,err
	}
	var things []model.Thing
	//err := st.db.Select(q.Eq("LocationId", locationId)).Find(&things)
	//return things, err
	for i := range site.Things {
		if site.Things[i].RoomID == int(locationId) {
			thing , err := r.GetThingById(model.ID(site.Things[i].RoomID))
			if err != nil {
				things = append(things, *thing)
			}
		}
	}
	return things, nil
}

func (VinculumRegistryStore) GetThingByIntegrationId(id string) (*model.Thing, error) {
	panic("implement me")
}

func (r *VinculumRegistryStore) GetAllDevices() ([]model.Device, error) {
	site , err := r.vApi.GetSite(true)
	if err != nil {
		return nil,err
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
			device.LocationId =  model.ID(*vDevices[i].Room)
		}

		if vDevices[i].ThingID != nil {
			device.ThingId =  model.ID(*vDevices[i].ThingID)
		}

		devices = append(devices, device)
	}
	return devices,nil
}

func (VinculumRegistryStore) GetDeviceById(Id model.ID) (*model.DeviceExtendedView, error) {
	panic("implement me")
}

func (VinculumRegistryStore) GetDevicesByLocationId(locationId model.ID) ([]model.Device, error) {
	panic("implement me")
}

func (VinculumRegistryStore) GetDeviceByIntegrationId(id string) (*model.Device, error) {
	panic("implement me")
}

func (VinculumRegistryStore) GetLocationByIntegrationId(id string) (*model.Location, error) {
	panic("implement me")
}

func (VinculumRegistryStore) UpsertThing(thing *model.Thing) (model.ID, error) {
	panic("implement me")
}

func (VinculumRegistryStore) UpsertService(service *model.Service) (model.ID, error) {
	panic("implement me")
}

func (VinculumRegistryStore) UpsertLocation(location *model.Location) (model.ID, error) {
	panic("implement me")
}

func (VinculumRegistryStore) DeleteThing(id model.ID) error {
	panic("implement me")
}

func (VinculumRegistryStore) DeleteService(id model.ID) error {
	panic("implement me")
}

func (VinculumRegistryStore) DeleteLocation(id model.ID) error {
	panic("implement me")
}

func (VinculumRegistryStore) ReindexAll() error {
	panic("implement me")
}

func (VinculumRegistryStore) ClearAll() error {
	panic("implement me")
}

func (r *VinculumRegistryStore) GetConnection() interface{} {
	return r
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
