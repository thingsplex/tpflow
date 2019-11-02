package fh

import (
	"github.com/pkg/errors"
	"github.com/thingsplex/tpflow/registry/model"
	"github.com/thingsplex/tpflow/utils"
)

type VinculumRegistryStoreBackend struct {
	things    []model.Thing
	services  []model.Service
	locations []model.Location
}

func NewVinculumRegistryStoreBackend(storeFile string) *VinculumRegistryStoreBackend {
	store := VinculumRegistryStoreBackend{}
	store.Connect()
	return &store
}

func (st *VinculumRegistryStoreBackend) Connect() error {
	return nil

}

func (st *VinculumRegistryStoreBackend) Disconnect() {
}

func (st *VinculumRegistryStoreBackend) GetThingById(Id model.ID) (*model.Thing, error) {
	return nil, nil
}

func (st *VinculumRegistryStoreBackend) GetServiceById(Id model.ID) (*model.Service, error) {
	return nil, nil
}

func (st *VinculumRegistryStoreBackend) GetServiceByFullAddress(address string) (*model.ServiceExtendedView, error) {
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

func (st *VinculumRegistryStoreBackend) GetLocationById(Id model.ID) (*model.Location, error) {
	//var location model.Location
	return nil, nil
}

func (st *VinculumRegistryStoreBackend) GetAllThings() ([]model.Thing, error) {
	return nil, nil
}

func (st *VinculumRegistryStoreBackend) ExtendThingsWithLocation(things []model.Thing) []model.ThingWithLocationView {
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

func (st *VinculumRegistryStoreBackend) GetAllServices() ([]model.Service, error) {
	//var services []Service
	//err := st.db.All(&services)
	//return services, err
	return nil, nil
}

// GetThingExtendedViewById return thing enhanced with linked services and location Alias
func (st *VinculumRegistryStoreBackend) GetThingExtendedViewById(Id model.ID) (*model.ThingExtendedView, error) {
	var thingExView model.ThingExtendedView
	//err := st.db.One("ID", Id, &thing)
	thingp, err := st.GetThingById(Id)

	thingExView.Thing = *thingp
	services, err := st.GetExtendedServices("", false, Id, model.IDnil)
	thingExView.Services = make([]model.ServiceExtendedView, len(services))
	for i := range services {
		thingExView.Services[i] = services[i]
	}
	location, _ := st.GetLocationById(thingp.LocationId)
	if location != nil {
		thingExView.LocationAlias = location.Alias
	}
	return &thingExView, err
}

func (st *VinculumRegistryStoreBackend) GetServiceByAddress(serviceName string, serviceAddress string) (*model.Service, error) {
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
func (st *VinculumRegistryStoreBackend) GetExtendedServices(serviceNameFilter string, filterWithoutAlias bool, thingIdFilter model.ID, locationIdFilter model.ID) ([]model.ServiceExtendedView, error) {
	var services []model.Service

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

func (st *VinculumRegistryStoreBackend) GetAllLocations() ([]model.Location, error) {
	//var locations []Location
	//err := st.db.All(&locations)
	//return locations, err
	return st.locations, nil
}

func (st *VinculumRegistryStoreBackend) GetThingByAddress(technology string, address string) (*model.Thing, error) {
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

func (st *VinculumRegistryStoreBackend) GetThingExtendedViewByAddress(technology string, address string) (*model.ThingExtendedView, error) {
	thing, err := st.GetThingByAddress(technology, address)
	if err != nil {
		return nil, err
	}
	return st.GetThingExtendedViewById(thing.ID)
}

func (st *VinculumRegistryStoreBackend) GetThingsByLocationId(locationId model.ID) ([]model.Thing, error) {
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

func (st *VinculumRegistryStoreBackend) GetThingByIntegrationId(id string) (*model.Thing, error) {
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

func (st *VinculumRegistryStoreBackend) GetLocationByIntegrationId(id string) (*model.Location, error) {
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



func (st *VinculumRegistryStoreBackend) UpsertThing(thing *model.Thing) (model.ID, error) {
	return thing.ID, nil
}

func (st *VinculumRegistryStoreBackend) UpsertService(service *model.Service) (model.ID, error) {

	return service.ID, nil
}

func (st *VinculumRegistryStoreBackend) UpsertLocation(location *model.Location) (model.ID, error) {
	return location.ID, nil
}

func (st *VinculumRegistryStoreBackend) DeleteThing(id model.ID) error {

	return nil
}

func (st *VinculumRegistryStoreBackend) DeleteService(id model.ID) error {

	return nil
}

func (st *VinculumRegistryStoreBackend) DeleteLocation(id model.ID) error {
	return nil
}


// Method to comply with Connector interface

func (st *VinculumRegistryStoreBackend) LoadConfig(config interface{}) error {
	return nil
}
func (st *VinculumRegistryStoreBackend) Init() error {
	return nil
}
func (st *VinculumRegistryStoreBackend) Stop() {

}
func (st *VinculumRegistryStoreBackend) GetConnection() interface{} {
	return &st

}
func (st *VinculumRegistryStoreBackend) GetState() string {
	return "RUNNING"
}
