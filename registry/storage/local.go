package storage

import (
	"encoding/gob"
	"github.com/thingsplex/tpflow/registry/model"

	"github.com/asdine/storm"
	gobcodec "github.com/asdine/storm/codec/gob"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/utils"
)

type LocalRegistryStore struct {
	thingRegistryStoreFile string
	db                     *storm.DB
	// in memory store
	things    []model.Thing
	services  []model.Service
	locations []model.Location
}

func (st *LocalRegistryStore) GetConfig() interface{} {
	return nil
}

func (st *LocalRegistryStore) SetDefaults() bool {
	return false
}

func NewThingRegistryStore(storeFile string) RegistryStorage {
	store := LocalRegistryStore{thingRegistryStoreFile: storeFile}
	store.Connect()
	return &store
}

func (st *LocalRegistryStore) GetServicesByLocationAndName(locId model.ID, name string) ([]model.Service, error) {
	return nil , nil
}

func (st *LocalRegistryStore) GetExtendedDevices() ([]model.DeviceExtendedView, error) {
	panic("implement me")
}

func (st *LocalRegistryStore)GetBackendName()string {
	return "local"
}

func (st *LocalRegistryStore) GetAllDevices() ([]model.Device, error) {
	panic("implement me")
}

func (st *LocalRegistryStore) GetDevicesByLocationId(locationId model.ID) ([]model.Device, error) {
	panic("implement me")
}

func (st *LocalRegistryStore) GetDevicesByThingId(locationId model.ID) ([]model.Device, error) {
	panic("implement me")
}

func (st *LocalRegistryStore) GetDeviceById(Id model.ID) (*model.DeviceExtendedView, error) {
	return nil,nil
}

func (st *LocalRegistryStore) GetDeviceByLocationId(locationId model.ID) ([]model.Device, error) {
	return nil,nil
}

func (st *LocalRegistryStore) GetDeviceByIntegrationId(id string) (*model.Device, error) {
	return nil,nil
}

func (st *LocalRegistryStore) Connect() error {
	var err error
	gob.Register([]interface{}{})
	st.db, err = storm.Open(st.thingRegistryStoreFile, storm.Codec(gobcodec.Codec))
	if err != nil {
		log.Error("<Reg> Can't open DB file . Error : ", err)
		return err
	}

	err = st.db.Init(&model.Thing{})
	if err != nil {
		log.Error("<Reg> Can't Init Things . Error : ", err)
		return err
	}
	// loading all things into memory
	err = st.db.All(&st.things)
	if err != nil {
		log.Error("<Reg> Can't Load Things . Error : ", err)
		return err
	}

	err = st.db.Init(&model.Service{})
	if err != nil {
		log.Error("<Reg> Can't Init Services . Error : ", err)
		return err
	}

	err = st.db.All(&st.services)
	if err != nil {
		log.Error("<Reg> Can't Load Services . Error : ", err)
		return err
	}

	// loading all services into memory

	err = st.db.Init(&model.Location{})
	if err != nil {
		log.Error("<Reg> Can't Init Things . Error : ", err)
		return err
	}

	err = st.db.All(&st.locations)

	return err

}

func (st *LocalRegistryStore) Disconnect() {
	st.db.Close()
}

func (st *LocalRegistryStore) GetThingById(Id model.ID) (*model.Thing, error) {
	//var thing Thing
	//err := st.db.One("ID", Id, &thing)
	for i := range st.things {
		if st.things[i].ID == Id {
			return &st.things[i], nil
		}
	}
	return nil, nil
}

func (st *LocalRegistryStore) GetServiceById(Id model.ID) (*model.Service, error) {
	//var service Service
	//err := st.db.One("ID", Id, &service)
	//return &service, err
	for i := range st.services {
		if st.services[i].ID == Id {
			return &st.services[i], nil
		}
	}
	return nil, nil
}

func (st *LocalRegistryStore) GetServiceByFullAddress(address string) (*model.ServiceExtendedView, error) {
	var serv model.ServiceExtendedView

	for i := range st.services {
		if utils.RouteIncludesTopic("+/+"+st.services[i].Address, address) {
			serv.Service = st.services[i]
			location, _ := st.GetLocationById(st.services[i].LocationId)
			if location != nil {
				serv.LocationAlias = location.Alias
				serv.LocationType = location.Type
				serv.LocationSubType = location.SubType
			}
			return &serv, nil
		}
	}
	return &serv, errors.New("Not found")
}

func (st *LocalRegistryStore) GetLocationById(Id model.ID) (*model.Location, error) {
	//var location Location
	//err := st.db.One("ID", Id, &location)
	//return &location, err
	for i := range st.locations {
		if st.locations[i].ID == Id {
			return &st.locations[i], nil
		}
	}
	return nil, nil
}

func (st *LocalRegistryStore) GetAllThings() ([]model.Thing, error) {
	//var things []Thing
	//err := st.db.All(&things)
	//return things, err
	return st.things, nil
}

func (st *LocalRegistryStore) ExtendThingsWithLocation(things []model.Thing) []model.ThingWithLocationView {
	response := make([]model.ThingWithLocationView, len(things))
	for i := range things {
		response[i].Thing = things[i]
		loc, _ := st.GetLocationById(things[i].LocationId)
		if loc != nil {
			response[i].LocationAlias = loc.Alias
		}

	}
	return response
}

func (st *LocalRegistryStore) GetAllServices() ([]model.Service, error) {
	//var services []Service
	//err := st.db.All(&services)
	//return services, err
	return st.services, nil
}

func (st *LocalRegistryStore) GetServicesByDeviceId(Id model.ID)([]model.Service,error) {
	return nil, errors.New("not implemented")
}

// GetThingExtendedViewById return thing enhanced with linked services and location Alias
func (st *LocalRegistryStore) GetThingExtendedViewById(Id model.ID) (*model.ThingExtendedView, error) {
	var thingExView model.ThingExtendedView
	//err := st.db.One("ID", Id, &thing)
	thingp, err := st.GetThingById(Id)

	thingExView.Thing = *thingp
	//services, err := st.GetExtendedServices("", false, Id, model.IDnil)
	//thingExView.Services = make([]model.ServiceExtendedView, len(services))
	//for i := range services {
	//	thingExView.Services[i] = services[i]
	//}
	location, _ := st.GetLocationById(thingp.LocationId)
	if location != nil {
		thingExView.LocationAlias = location.Alias
	}
	return &thingExView, err
}

func (st *LocalRegistryStore) GetServiceByAddress(serviceName string, serviceAddress string) (*model.Service, error) {
	//service := Service{}
	//err := st.db.Select(q.And(q.Eq("Name", serviceName), q.Eq("Address", serviceAddress))).First(&service)
	for i := range st.services {
		if st.services[i].Name == serviceName && st.services[i].Address == serviceAddress {
			return &st.services[i], nil
		}
	}
	return nil, errors.New("Not found")
}

// GetExtendedServices return services enhanced with location Alias
func (st *LocalRegistryStore) GetExtendedServices(serviceNameFilter string, filterWithoutAlias bool, thingIdFilter model.ID, locationIdFilter model.ID) ([]model.ServiceExtendedView, error) {
	var services []model.Service
	//var matcher []q.Matcher
	//if serviceNameFilter != "" {
	//	match := q.Eq("Name", serviceNameFilter)
	//	matcher = append(matcher, match)
	//}
	//if locationIdFilter != IDnil {
	//	match := q.Eq("LocationId", locationIdFilter)
	//	matcher = append(matcher, match)
	//}
	//if filterWithoutAlias {
	//	match := q.Not(q.Eq("Alias", ""))
	//	matcher = append(matcher, match)
	//}
	//if thingIdFilter != IDnil {
	//	matcher = append(matcher, q.Eq("ParentContainerId", thingIdFilter))
	//	matcher = append(matcher, q.Eq("ParentContainerType", ThingContainer))
	//}
	for i := range st.services {
		if serviceNameFilter != "" {
			if st.services[i].Name != serviceNameFilter {
				continue
			}
		}

		if locationIdFilter != model.IDnil {
			if st.services[i].LocationId != locationIdFilter {
				continue
			}
		}

		if filterWithoutAlias {
			if st.services[i].Alias == "" {
				continue
			}
		}

		if thingIdFilter != model.IDnil {
			if st.services[i].ParentContainerId != thingIdFilter || st.services[i].ParentContainerType != model.ThingContainer {
				continue
			}
		}
		services = append(services, st.services[i])

	}

	//err := st.db.Select(matcher...).Find(&services)
	//if err != nil {
	//	log.Error("<Reg> Can't fetch services . Error : ", err)
	//	return nil, err
	//}
	var result []model.ServiceExtendedView
	for si := range services {
		serviceResponse := model.ServiceExtendedView{Service: services[si]}
		location, _ := st.GetLocationById(serviceResponse.LocationId)
		if location != nil {
			serviceResponse.LocationAlias = location.Alias
		}
		result = append(result, serviceResponse)
	}
	return result, nil
}

func (st *LocalRegistryStore) GetAllLocations() ([]model.Location, error) {
	//var locations []Location
	//err := st.db.All(&locations)
	//return locations, err
	return st.locations, nil
}

func (st *LocalRegistryStore) GetThingByAddress(technology string, address string) (*model.Thing, error) {
	//var thing Thing
	//err := st.db.Select(q.And(q.Eq("Address", address), q.Eq("CommTechnology", technology))).First(&thing)
	//return &thing,err
	for i := range st.things {
		if st.things[i].Address == address && st.things[i].CommTechnology == technology {
			return &st.things[i], nil
		}
	}
	return nil, errors.New("Not found")
}

func (st *LocalRegistryStore) GetThingExtendedViewByAddress(technology string, address string) (*model.ThingExtendedView, error) {
	thing, err := st.GetThingByAddress(technology, address)
	if err != nil {
		return nil, err
	}
	return st.GetThingExtendedViewById(thing.ID)
}

func (st *LocalRegistryStore) GetThingsByLocationId(locationId model.ID) ([]model.Thing, error) {
	var things []model.Thing
	//err := st.db.Select(q.Eq("LocationId", locationId)).Find(&things)
	//return things, err
	for i := range st.things {
		if st.things[i].LocationId == locationId {
			things = append(things, st.things[i])
		}
	}
	return things, nil
}

func (st *LocalRegistryStore) GetThingByIntegrationId(id string) (*model.Thing, error) {
	//var thing Thing
	//err := st.db.Select(q.Eq("IntegrationId", id)).First(&thing)
	//return &thing, err
	for i := range st.things {
		if st.things[i].IntegrationId == id {
			return &st.things[i], nil
		}
	}
	return nil, errors.New("Not found")
}

func (st *LocalRegistryStore) GetLocationByIntegrationId(id string) (*model.Location, error) {
	//var location Location
	//err := st.db.Select(q.Eq("IntegrationId", id)).First(&location)
	//return &location, err

	for i := range st.locations {
		if st.locations[i].IntegrationId == id {
			return &st.locations[i], nil
		}
	}
	return nil, errors.New("Not found")
}

//func (st *LocalRegistryStore) GetFlatInterfaces(thingAddr string, thingTech string, serviceName string, intfMsgType string, locationId ID, thingId ID) ([]InterfaceFlatView, error) {
//	var result []InterfaceFlatView
//	//things, err  := st.GetAllThings()
//	var things []Thing
//	var matcher []q.Matcher
//	if thingAddr != "" {
//		match := q.Eq("Address", thingAddr)
//		matcher = append(matcher, match)
//	}
//	if thingTech != "" {
//		match := q.Eq("CommTechnology", thingTech)
//		matcher = append(matcher, match)
//	}
//	if thingId != 0 {
//		match := q.Eq("ID", thingId)
//		matcher = append(matcher, match)
//	}
//	err := st.db.Select(matcher...).Find(&things)
//	if err != nil {
//		return nil, err
//	}
//	for thi := range things {
//		for si := range things[thi].Services {
//			for inti := range things[thi].Services[si].Interfaces {
//				if (serviceName == "" || things[thi].Services[si].Name == serviceName) &&
//					(intfMsgType == "" || things[thi].Services[si].Interfaces[inti].MsgType == intfMsgType) &&
//					(locationId == 0 || things[thi].Services[si].LocationId == locationId) {
//
//					flatIntf := InterfaceFlatView{}
//					flatIntf.ThingId = things[thi].ID
//					flatIntf.ThingAddress = things[thi].Address
//					flatIntf.ThingTech = things[thi].CommTechnology
//					flatIntf.ThingAlias = things[thi].Alias
//					flatIntf.ServiceId = things[thi].Services[si].ID
//					flatIntf.ServiceName = things[thi].Services[si].Name
//					flatIntf.ServiceAlias = things[thi].Services[si].Alias
//					flatIntf.ServiceAddress = things[thi].Services[si].Address
//					flatIntf.InterfaceType = things[thi].Services[si].Interfaces[inti].Type
//					flatIntf.InterfaceMsgType = things[thi].Services[si].Interfaces[inti].MsgType
//					flatIntf.InterfaceValueType = things[thi].Services[si].Interfaces[inti].ValueContainer
//
//					//pt:j1/mt:evt/rt:dev/rn:zw/ad:1/sv:meter_elec/ad:21_0
//					prefix := "pt:j1/mt:evt"
//					if strings.Contains(prefix+things[thi].Services[si].Interfaces[inti].MsgType, "cmd") {
//						prefix = "pt:j1/mt:cmd"
//					}
//					flatIntf.InterfaceAddress = prefix + things[thi].Services[si].Address
//					location, _ := st.GetLocationById(things[thi].Services[si].LocationId)
//					if location != nil {
//						flatIntf.LocationId = location.ID
//						flatIntf.LocationAlias = location.Alias
//						flatIntf.LocationType = location.Type
//
//					}
//					location, _ = st.GetLocationById(things[thi].LocationId)
//					if location != nil {
//						if location.Alias != flatIntf.LocationAlias {
//							flatIntf.LocationAlias = location.Alias + " " + flatIntf.LocationAlias
//						}
//						if flatIntf.LocationType == "" {
//							flatIntf.LocationType = location.Type
//						}
//					}
//
//					result = append(result, flatIntf)
//				}
//
//			}
//		}
//	}
//	return result, nil
//}

func (st *LocalRegistryStore) UpsertThing(thing *model.Thing) (model.ID, error) {
	var err error
	if thing.ID == model.IDnil {
		err = st.db.Save(thing)
		if err == nil {
			st.things = append(st.things, *thing)
		}

	} else {
		err = st.db.Update(thing)
		if err == nil {
			th, _ := st.GetThingById(thing.ID)
			mergo.Merge(th, *thing, mergo.WithOverride)

		}

	}

	if err != nil {
		log.Error("<Reg> Can't save thing . Error :", err)
		return 0, err
	} else {
		log.Debug("<Reg> Thing saved ")
	}

	// Updating linked services
	services, _ := st.GetExtendedServices("", false, thing.ID, model.IDnil)
	var isChanged bool
	for i := range services {
		isChanged = false
		//if services[i].Alias == "" {
		services[i].Alias = thing.Alias
		isChanged = true
		//}
		//if services[i].LocationId == IDnil {
		services[i].LocationId = thing.LocationId
		//isChanged = true
		//}
		if isChanged {
			st.UpsertService(&services[i].Service)
		}
	}
	return thing.ID, nil
}

func (st *LocalRegistryStore) UpsertService(service *model.Service) (model.ID, error) {
	var err error
	// Check if service is already registered in system , if record already exits , updating the record
	if service.ID == model.IDnil {
		serviceCheck, err := st.GetServiceByAddress(service.Name, service.Address)
		//serviceCheck := Service{}
		//err = st.db.Select(q.And(q.Eq("Name", service.Name), q.Eq("Address", service.Address))).First(&serviceCheck)
		if err == nil {
			service.ID = serviceCheck.ID
		}
	}
	if service.ID == model.IDnil {
		// Create new service
		err = st.db.Save(service)
		if err == nil {
			st.services = append(st.services, *service)
		}
	} else {
		err = st.db.Update(service)
		if err == nil {
			sv, _ := st.GetServiceByAddress(service.Name, service.Address)
			mergo.Merge(sv, *service, mergo.WithOverride)
		}

	}

	if err != nil {
		log.Error("<Reg> Can't save service . Error :", err)
		return 0, err
	} else {
		log.Debug("<Reg> Service saved ")
	}

	return service.ID, nil
}

func (st *LocalRegistryStore) UpsertLocation(location *model.Location) (model.ID, error) {
	var err error
	if location.ID == 0 {
		err = st.db.Save(location)
		if err == nil {
			st.locations = append(st.locations, *location)
		}
	} else {
		err = st.db.Update(location)
		if err == nil {
			loc, _ := st.GetLocationById(location.ID)
			*loc = *location
		}

	}

	if err != nil {
		log.Error("Can't save location . Error :", err)
		return 0, err
	} else {
		log.Debug("Location saved ")
	}

	return location.ID, nil
}

func (st *LocalRegistryStore) DeleteThing(id model.ID) error {
	thing, err := st.GetThingById(id)
	log.Debug("<Reg> Deleting thing ", thing.ID)
	if err != nil {
		return err
	}
	st.db.DeleteStruct(thing)
	// Deleting all linked services
	services, _ := st.GetExtendedServices("", false, id, model.IDnil)
	var servIDs []model.ID
	for i := range services {
		servIDs = append(servIDs, services[i].ID)
	}
	for _, id := range servIDs {
		st.DeleteService(id)
	}

	for i := range st.things {
		if st.things[i].ID == id {
			st.things = append(st.things[:i], st.things[i+1:]...)
			break
		}
	}

	return nil
}

func (st *LocalRegistryStore) DeleteService(id model.ID) error {
	service, err := st.GetServiceById(id)
	log.Debug("<Reg> Deleting service = ", service.ID)
	if err != nil {
		return err
	}
	err = st.db.DeleteStruct(service)

	if err == nil {
		for i := range st.services {
			if st.services[i].ID == id {
				st.services = append(st.services[:i], st.services[i+1:]...)
				break
			}
		}
	}
	return err
}

func (st *LocalRegistryStore) DeleteLocation(id model.ID) error {
	location, err := st.GetLocationById(id)
	log.Debug("<Reg> Deleting location = ", location.ID)
	if err != nil {
		return err
	}
	err = st.db.DeleteStruct(location)
	if err == nil {
		for i := range st.locations {
			if st.locations[i].ID == id {
				st.locations = append(st.locations[:i], st.locations[i+1:]...)
			}
		}
	}
	return nil
}

func (st *LocalRegistryStore) ReindexAll() error {
	log.Info("Starting reindex")
	err := st.db.ReIndex(&model.Thing{})
	err = st.db.ReIndex(&model.Location{})
	err = st.db.ReIndex(&model.Service{})
	log.Info("Reindex is complete")
	return err
}

func (st *LocalRegistryStore) ClearAll() error {
	thing := model.Thing{}
	location := model.Location{}
	service := model.Service{}
	st.db.Drop(thing)
	st.db.Drop(location)
	st.db.Drop(service)
	st.locations = nil
	st.services = nil
	st.things = nil

	err := st.db.Init(&thing)
	if err != nil {
		log.Error("<Reg> Can't Init Things . Error : ", err)
		return err
	}

	err = st.db.Init(&location)
	if err != nil {
		log.Error("<Reg> Can't Init Locations . Error : ", err)
		return err
	}

	err = st.db.Init(&service)
	if err != nil {
		log.Error("<Reg> Can't Init Service . Error : ", err)
		return err
	}
	return nil
}

func (st *LocalRegistryStore) Sync() error {
	log.Info("<Reg> Sync is not implemented")
	return nil
}

// Method to comply with Connector interface

func (st *LocalRegistryStore) LoadConfig(config interface{}) error {
	return nil
}
func (st *LocalRegistryStore) Init() error {
	return nil
}
func (st *LocalRegistryStore) Stop() {

}
func (st *LocalRegistryStore) GetConnection() interface{} {
	return &st

}
func (st *LocalRegistryStore) GetState() string {
	return "RUNNING"
}
