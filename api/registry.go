package api

import (
	"encoding/json"
	"fmt"
	"github.com/futurehomeno/fimpgo"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/registry/model"
	"github.com/thingsplex/tpflow/registry/storage"
	"net/http"
	"strconv"
)

type RegistryApi struct {
	reg  storage.RegistryStorage
	echo *echo.Echo
	msgTransport *fimpgo.MqttTransport
}

func NewRegistryApi(ctx storage.RegistryStorage, echo *echo.Echo) *RegistryApi {
	ctxApi := RegistryApi{reg: ctx, echo: echo}
	//ctxApi.RegisterRestApi()
	return &ctxApi
}

func (api *RegistryApi) RegisterRestApi() {
	api.echo.GET("/fimp/api/registry/things", func(c echo.Context) error {

		var things []model.Thing
		var locationId int
		var err error
		locationIdStr := c.QueryParam("locationId")
		locationId, _ = strconv.Atoi(locationIdStr)

		if locationId != 0 {
			things, err = api.reg.GetThingsByLocationId(model.ID(locationId))
		} else {
			things, err = api.reg.GetAllThings()
		}
		thingsWithLocation := api.reg.ExtendThingsWithLocation(things)
		if err == nil {
			return c.JSON(http.StatusOK, thingsWithLocation)
		} else {
			return c.JSON(http.StatusInternalServerError, err)
		}

	})

	api.echo.GET("/fimp/api/registry/services", func(c echo.Context) error {
		serviceName := c.QueryParam("serviceName")
		locationIdStr := c.QueryParam("locationId")
		thingIdStr := c.QueryParam("thingId")
		thingId, _ := strconv.Atoi(thingIdStr)
		locationId, _ := strconv.Atoi(locationIdStr)
		filterWithoutAliasStr := c.QueryParam("filterWithoutAlias")
		var filterWithoutAlias bool
		if filterWithoutAliasStr == "true" {
			filterWithoutAlias = true
		}
		services, err := api.reg.GetExtendedServices(serviceName, filterWithoutAlias, model.ID(thingId), model.ID(locationId))

		if err == nil {
			return c.JSON(http.StatusOK, services)
		} else {
			return c.JSON(http.StatusInternalServerError, err)
		}
	})

	api.echo.GET("/fimp/api/registry/service", func(c echo.Context) error {
		serviceAddress := c.QueryParam("address")
		log.Info("<REST> Service search , address =  ", serviceAddress)
		services, err := api.reg.GetServiceByFullAddress(serviceAddress)
		if err == nil {
			return c.JSON(http.StatusOK, services)
		} else {
			return c.JSON(http.StatusInternalServerError, err)
		}
	})

	api.echo.PUT("/fimp/api/registry/service", func(c echo.Context) error {
		service := model.Service{}
		err := c.Bind(&service)
		if err == nil {
			log.Info("<REST> Saving service")
			api.reg.UpsertService(&service)
			return c.NoContent(http.StatusOK)
		} else {
			log.Info("<REST> Can't bind service")
			return c.JSON(http.StatusInternalServerError, err)
		}
	})

	api.echo.PUT("/fimp/api/registry/location", func(c echo.Context) error {
		location := model.Location{}
		err := c.Bind(&location)
		if err == nil {
			log.Info("<REST> Saving location")
			api.reg.UpsertLocation(&location)
			return c.NoContent(http.StatusOK)
		} else {
			log.Info("<REST> Can't bind location")
			return c.JSON(http.StatusInternalServerError, err)
		}
	})

	api.echo.GET("/fimp/api/registry/interfaces", func(c echo.Context) error {
		var err error
		//thingAddr := c.QueryParam("thingAddr")
		//thingTech := c.QueryParam("thingTech")
		//serviceName := c.QueryParam("serviceName")
		//intfMsgType := c.QueryParam("intfMsgType")
		//locationIdStr := c.QueryParam("locationId")
		//var locationId int
		//locationId, _ = strconv.Atoi(locationIdStr)
		//var thingId int
		//thingIdStr := c.QueryParam("thingId")
		//thingId, _ = strconv.Atoi(thingIdStr)
		//services, err := thingRegistryStore.GetFlatInterfaces(thingAddr, thingTech, serviceName, intfMsgType, registry.ID(locationId), registry.ID(thingId))
		services := []model.ServiceExtendedView{}
		if err == nil {
			return c.JSON(http.StatusOK, services)
		} else {
			return c.JSON(http.StatusInternalServerError, err)
		}
	})

	api.echo.GET("/fimp/api/registry/locations", func(c echo.Context) error {
		locations, err := api.reg.GetAllLocations()
		if err == nil {
			return c.JSON(http.StatusOK, locations)
		} else {
			return c.JSON(http.StatusInternalServerError, err)
		}
	})

	api.echo.GET("/fimp/api/registry/thing/:tech/:address", func(c echo.Context) error {
		things, err := api.reg.GetThingExtendedViewByAddress(c.Param("tech"), c.Param("address"))
		if err == nil {
			return c.JSON(http.StatusOK, things)
		} else {
			return c.JSON(http.StatusInternalServerError, err)
		}

	})
	api.echo.DELETE("/fimp/api/registry/clear_all", func(c echo.Context) error {
		api.reg.ClearAll()
		return c.NoContent(http.StatusOK)
	})

	api.echo.POST("/fimp/api/registry/reindex", func(c echo.Context) error {
		api.reg.ReindexAll()
		return c.NoContent(http.StatusOK)
	})

	api.echo.PUT("/fimp/api/registry/thing", func(c echo.Context) error {
		thing := model.Thing{}
		err := c.Bind(&thing)
		fmt.Println(err)
		if err == nil {
			log.Info("<REST> Saving thing")
			api.reg.UpsertThing(&thing)
			return c.NoContent(http.StatusOK)
		} else {
			log.Info("<REST> Can't bind thing")
			return c.JSON(http.StatusInternalServerError, err)
		}
		return c.NoContent(http.StatusOK)
	})

	api.echo.DELETE("/fimp/api/registry/thing/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		thingId, _ := strconv.Atoi(idStr)
		err := api.reg.DeleteThing(model.ID(thingId))
		if err == nil {
			return c.NoContent(http.StatusOK)
		}
		log.Error("<REST> Can't delete thing ")
		return c.JSON(http.StatusInternalServerError, err)
	})

	api.echo.DELETE("/fimp/api/registry/location/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		thingId, _ := strconv.Atoi(idStr)
		err := api.reg.DeleteLocation(model.ID(thingId))
		if err == nil {
			return c.NoContent(http.StatusOK)
		}
		log.Error("<REST> Failed to delete thing . Error : ", err)
		return c.JSON(http.StatusInternalServerError, err)
	})

}

func (api *RegistryApi) RegisterMqttApi(msgTransport *fimpgo.MqttTransport) {
	api.msgTransport = msgTransport
	api.msgTransport.Subscribe("pt:j1/mt:cmd/rt:app/rn:registry/ad:1")
	apiCh := make(fimpgo.MessageCh, 10)
	api.msgTransport.RegisterChannel("registry-api",apiCh)
	var fimp *fimpgo.FimpMessage
	go func() {
		for {

			newMsg := <-apiCh
			log.Debug("New registry message of type ", newMsg.Payload.Type)
			var err error
			fimp = nil
			switch newMsg.Payload.Type {
			case "cmd.registry.get_things":
				var things []model.Thing
				var locationId int

				val,_ := newMsg.Payload.GetStrMapValue()
				locationIdStr,_ := val["location_id"]

				locationId, _ = strconv.Atoi(locationIdStr)

				if locationId != 0 {
					things, err = api.reg.GetThingsByLocationId(model.ID(locationId))
				} else {
					things, err = api.reg.GetAllThings()
				}
				if err != nil {
					log.Error("<RegApi> can't get things from registry err:",err)
					break
				}
				thingsWithLocation := api.reg.ExtendThingsWithLocation(things)
				fimp = fimpgo.NewMessage("evt.registry.things_report", "tpflow", "object", thingsWithLocation, nil, nil, newMsg.Payload)

			case "cmd.registry.get_devices":
				var devices []model.Device
				var locationId int

				val,_ := newMsg.Payload.GetStrMapValue()
				locationIdStr,_ := val["location_id"]

				locationId, _ = strconv.Atoi(locationIdStr)

				if locationId != 0 {
					devices, err = api.reg.GetDevicesByLocationId(model.ID(locationId))
				} else {
					devices, err = api.reg.GetAllDevices()
				}
				if err != nil {
					log.Error("<RegApi> can't get things from registry err:",err)
					break
				}
				//thingsWithLocation := api.reg.ExtendDevicesWithLocation(things)
				fimp = fimpgo.NewMessage("evt.registry.things_report", "tpflow", "object", devices, nil, nil, newMsg.Payload)


			case "cmd.registry.get_services":
				log.Debug("Getting services ")
				//val,err := newMsg.Payload.GetStrMapValue()
				//if err != nil {
				//	fimp = nil
				//	log.Debug("Error while requesting services . Error :",err)
				//	break
				//}
				//serviceName,_ := val["service_name"]
				//locationIdStr,_ := val["location_id"]
				//thingIdStr,_ := val["thing_id"]
				//thingId, _ := strconv.Atoi(thingIdStr)
				//locationId, _ := strconv.Atoi(locationIdStr)
				//filterWithoutAliasStr,_ := val["filter_without_alias"]
				//var filterWithoutAlias bool
				//if filterWithoutAliasStr == "true" {
				//	filterWithoutAlias = true
				//}
				log.Debug("Getting extended services ")
				//services, err := api.reg.GetExtendedServices(serviceName, filterWithoutAlias, model.ID(thingId), model.ID(locationId))
				services, err := api.reg.GetAllServices()


				if err == nil {
					fimp = fimpgo.NewMessage("evt.registry.services_report", "tpflow", "object", services, nil, nil, newMsg.Payload)
				} else {
					log.Error("<RegApi> Can't get list of extended services . Error :",err)
					fimp = nil
				}

			case "cmd.registry.get_service":
				val,err := newMsg.Payload.GetStrMapValue()
				serviceAddress,_ := val["address"]
				log.Info("<RegApi> Service search , address =  ", serviceAddress)
				services, err := api.reg.GetServiceByFullAddress(serviceAddress)
				if err == nil {
					fimp = fimpgo.NewMessage("evt.registry.service_report", "object", "string", services, nil, nil, newMsg.Payload)
				} else {
					log.Error("<RegApi> Can't get service info . Error :",err)
					fimp = nil
				}

			case "cmd.registry.update_service":
				service := model.Service{}
				serviceJsonDef := newMsg.Payload.GetRawObjectValue()
				err := json.Unmarshal(serviceJsonDef, &service)
				result := make(map[string]string)
				result["status"] = ""
				result["id"] = ""
				if err != nil {
					log.Error("<RegApi> Can't unmarshal  service.")
					result["status"] = "error"
					break
				}
				id , err := api.reg.UpsertService(&service)
				if err == nil {
					result["id"] = string(id)
					result["status"] = "ok"
				}

				fimp = fimpgo.NewStrMapMessage("evt.registry.update_service_report", "tpflow", result, nil, nil, newMsg.Payload)

			case "cmd.registry.update_thing":
				thing := model.Thing{}
				thingJsonDef := newMsg.Payload.GetRawObjectValue()
				err := json.Unmarshal(thingJsonDef, &thing)
				result := make(map[string]string)
				result["status"] = ""
				result["id"] = ""
				if err != nil {
					log.Error("<RegApi> Can't unmarshal  service.")
					result["status"] = "error"
					break
				}
				id , err := api.reg.UpsertThing(&thing)
				if err == nil {
					result["id"] = string(id)
					result["status"] = "ok"
				}

				fimp = fimpgo.NewStrMapMessage("evt.registry.update_thing_report", "tpflow", result, nil, nil, newMsg.Payload)

			case "cmd.registry.update_location":
				location := model.Location{}
				locationJsonDef := newMsg.Payload.GetRawObjectValue()
				err := json.Unmarshal(locationJsonDef, &location)
				result := make(map[string]string)
				result["status"] = ""
				result["id"] = ""
				if err != nil {
					log.Error("<RegApi> Can't unmarshal location.")
					break
				}
				id , err := api.reg.UpsertLocation(&location)
				if err == nil {
					result["id"] = string(id)
					result["status"] = "ok"
				}
				fimp = fimpgo.NewStrMapMessage("evt.registry.update_location_report", "tpflow", result, nil, nil, newMsg.Payload)

			case "cmd.registry.get_locations":
				locations, err := api.reg.GetAllLocations()
				if err != nil {
					log.Error("<RegApi> Can't get locations .Error:",err)
				}
				fimp = fimpgo.NewMessage("evt.registry.locations_report", "tpflow", "object", locations, nil, nil, newMsg.Payload)

			case "cmd.registry.get_thing":
				val,_ := newMsg.Payload.GetStrMapValue()
				tech , _ := val["tech"]
				address , _ := val["address"]

				thing, err := api.reg.GetThingExtendedViewByAddress(tech,address)
				if err != nil {
					log.Error("<RegApi> Can't get thing .Error:",err)
				}
				fimp = fimpgo.NewMessage("evt.registry.thing_report", "tpflow", "object", thing, nil, nil, newMsg.Payload)

			case "cmd.registry.delete_thing":
				idStr , _ := newMsg.Payload.GetStringValue()
				thingId, _ := strconv.Atoi(idStr)
				err := api.reg.DeleteThing(model.ID(thingId))
				var resp string
				if err != nil {
					resp = err.Error()
				}
				fimp = fimpgo.NewMessage("evt.registry.delete_thing_report", "tpflow", "string", resp, nil, nil, newMsg.Payload)

			case "cmd.registry.delete_location":
				idStr , _ := newMsg.Payload.GetStringValue()
				thingId, _ := strconv.Atoi(idStr)
				err := api.reg.DeleteLocation(model.ID(thingId))
				var resp string
				if err != nil {
					resp = err.Error()
				}
				fimp = fimpgo.NewMessage("evt.registry.delete_location_report", "tpflow", "string", resp, nil, nil, newMsg.Payload)

			case "cmd.registry.factory_reset":
				log.Info("<RegApi> Registry FACTORY RESET")
				api.reg.ClearAll()
			}
			if fimp != nil {
				if err := api.msgTransport.RespondToRequest(newMsg.Payload, fimp); err != nil {
					adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeApp, ResourceName: "tpflow", ResourceAddress: "1",}
					api.msgTransport.Publish(&adr, fimp)
				}
				log.Debug(err)
			}else {
				//log.Error("<reg-api> Error , nothing to return . Err:",err)
			}

		}
	}()


}
