package api

import (
	"fmt"
	"github.com/alivinco/tpflow/registry"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"net/http"
	"strconv"
)

type RegistryApi struct {
	reg  *registry.ThingRegistryStore
	echo *echo.Echo
}



func NewRegistryApi(ctx *registry.ThingRegistryStore,echo *echo.Echo) *RegistryApi {
	ctxApi := RegistryApi{reg:ctx,echo:echo}
	ctxApi.RegisterRestApi()
	return &ctxApi
}

func (api * RegistryApi) RegisterRestApi() {
	api.echo.GET("/fimp/api/registry/things", func(c echo.Context) error {

		var things []registry.Thing
		var locationId int
		var err error
		locationIdStr := c.QueryParam("locationId")
		locationId, _ = strconv.Atoi(locationIdStr)

		if locationId != 0 {
			things, err = api.reg.GetThingsByLocationId(registry.ID(locationId))
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
		locationId , _ := strconv.Atoi(locationIdStr)
		filterWithoutAliasStr:= c.QueryParam("filterWithoutAlias")
		var filterWithoutAlias bool
		if filterWithoutAliasStr == "true" {
			filterWithoutAlias = true
		}
		services, err := api.reg.GetExtendedServices(serviceName,filterWithoutAlias,registry.ID(thingId),registry.ID(locationId))
		if err == nil {
			return c.JSON(http.StatusOK, services)
		} else {
			return c.JSON(http.StatusInternalServerError, err)
		}
	})

	api.echo.GET("/fimp/api/registry/service", func(c echo.Context) error {
		serviceAddress := c.QueryParam("address")
		log.Info("<REST> Service search , address =  ",serviceAddress)
		services, err := api.reg.GetServiceByFullAddress(serviceAddress)
		if err == nil {
			return c.JSON(http.StatusOK, services)
		} else {
			return c.JSON(http.StatusInternalServerError, err)
		}
	})

	api.echo.PUT("/fimp/api/registry/service", func(c echo.Context) error {
		service := registry.Service{}
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
		location := registry.Location{}
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
		services := []registry.ServiceExtendedView{}
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
		thing := registry.Thing{}
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
		err := api.reg.DeleteThing(registry.ID(thingId))
		if err == nil {
			return c.NoContent(http.StatusOK)
		}
		log.Error("<REST> Can't delete thing ")
		return c.JSON(http.StatusInternalServerError, err)
	})

	api.echo.DELETE("/fimp/api/registry/location/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		thingId, _ := strconv.Atoi(idStr)
		err := api.reg.DeleteLocation(registry.ID(thingId))
		if err == nil {
			return c.NoContent(http.StatusOK)
		}
		log.Error("<REST> Failed to delete thing . Error : ",err)
		return c.JSON(http.StatusInternalServerError, err)
	})


}
