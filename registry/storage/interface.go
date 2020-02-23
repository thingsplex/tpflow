package storage

import "github.com/thingsplex/tpflow/registry/model"

type RegistryStorage interface {
	Connect() error
	Disconnect()
	GetServiceById(Id model.ID) (*model.Service, error)
	GetServiceByFullAddress(address string) (*model.ServiceExtendedView, error)
	GetLocationById(Id model.ID) (*model.Location, error)
	GetAllThings() ([]model.Thing, error)
	ExtendThingsWithLocation(things []model.Thing) []model.ThingWithLocationView
	GetAllServices() ([]model.Service, error)
	GetThingExtendedViewById(Id model.ID) (*model.ThingExtendedView, error)
	GetServiceByAddress(serviceName string, serviceAddress string) (*model.Service, error)
	GetExtendedServices(serviceNameFilter string, filterWithoutAlias bool, thingIdFilter model.ID, locationIdFilter model.ID) ([]model.ServiceExtendedView, error)
	GetAllLocations() ([]model.Location, error)
	GetThingById(Id model.ID) (*model.Thing, error)
	GetThingByAddress(technology string, address string) (*model.Thing, error)
	GetThingExtendedViewByAddress(technology string, address string) (*model.ThingExtendedView, error)
	GetThingsByLocationId(locationId model.ID) ([]model.Thing, error)
	GetThingByIntegrationId(id string) (*model.Thing, error)
	GetAllDevices() ([]model.Device, error)
	GetDeviceById(Id model.ID) (*model.DeviceExtendedView, error)
	GetDevicesByLocationId(locationId model.ID) ([]model.Device, error)
	GetDevicesByThingId(locationId model.ID) ([]model.Device, error)
	GetDeviceByIntegrationId(id string) (*model.Device, error)
	GetLocationByIntegrationId(id string) (*model.Location, error)
	UpsertThing(thing *model.Thing) (model.ID, error)
	UpsertService(service *model.Service) (model.ID, error)
	UpsertLocation(location *model.Location) (model.ID, error)
	DeleteThing(id model.ID) error
	DeleteService(id model.ID) error
	DeleteLocation(id model.ID) error
	ReindexAll() error
	ClearAll() error

	LoadConfig(config interface{}) error
	Init() error
	Stop()
	GetConnection() interface{}
	GetState() string
}
