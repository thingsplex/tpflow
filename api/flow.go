package api

import (
	"encoding/json"
	"github.com/alivinco/tpflow/connector/plugins"
	"github.com/alivinco/tpflow/flow"
	"github.com/alivinco/tpflow/model"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net/http"
)

type FlowApi struct {
	flowManager *flow.Manager
	echo        *echo.Echo
}

func NewFlowApi(flowManager *flow.Manager, echo *echo.Echo) *FlowApi {
	ctxApi := FlowApi{flowManager: flowManager, echo: echo}
	ctxApi.RegisterRestApi()
	return &ctxApi
}

func (ctx *FlowApi) RegisterRestApi() {
	ctx.echo.GET("/fimp/flow/list", func(c echo.Context) error {
		resp := ctx.flowManager.GetFlowList()
		return c.JSON(http.StatusOK, resp)
	})
	ctx.echo.GET("/fimp/flow/definition/:id", func(c echo.Context) error {
		id := c.Param("id")
		var resp *model.FlowMeta
		if id == "-" {
			flow := ctx.flowManager.GenerateNewFlow()
			resp = &flow
		} else {
			resp = ctx.flowManager.GetFlowById(id).FlowMeta
		}

		return c.JSON(http.StatusOK, resp)
	})

	ctx.echo.GET("/fimp/connector/template/:id", func(c echo.Context) error {
		id := c.Param("id")
		result := plugins.GetConfigurationTemplate(id)
		return c.JSON(http.StatusOK, result)
	})

	ctx.echo.GET("/fimp/connector/plugins", func(c echo.Context) error {
		result := plugins.GetPlugins()
		return c.JSON(http.StatusOK, result)
	})

	ctx.echo.GET("/fimp/connector/list", func(c echo.Context) error {
		result := ctx.flowManager.GetConnectorRegistry().GetAllInstances()
		return c.JSON(http.StatusOK, result)
	})

	ctx.echo.POST("/fimp/flow/definition/:id", func(c echo.Context) error {
		id := c.Param("id")
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		ctx.flowManager.UpdateFlowFromJsonAndSaveToStorage(id, body)
		return c.NoContent(http.StatusOK)
	})

	ctx.echo.PUT("/fimp/flow/definition/import", func(c echo.Context) error {
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		ctx.flowManager.ImportFlow(body)
		return c.NoContent(http.StatusOK)
	})

	ctx.echo.PUT("/fimp/flow/definition/import_from_url", func(c echo.Context) error {

		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		request := ImportFlowFromUrlRequest{}
		err = json.Unmarshal(body, &request)
		if err != nil {
			log.Error("Can't parse request ", err)
		}

		// Get the data
		resp, err := http.Get(request.Url)
		if err != nil {
			return err
		}
		flow, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error("Can't read file from url ", err)
			return err
		}
		log.Info("Importing flow")
		ctx.flowManager.ImportFlow(flow)
		return c.NoContent(http.StatusOK)
	})

	ctx.echo.POST("/fimp/flow/ctrl/:id/:op", func(c echo.Context) error {
		id := c.Param("id")
		op := c.Param("op")

		switch op {
		case "send-inclusion-report":
			ctx.flowManager.GetFlowById(id).SendInclusionReport()
		case "send-exclusion-report":
			ctx.flowManager.GetFlowById(id).SendExclusionReport()
		case "start":
			ctx.flowManager.ControlFlow("START", id)
		case "stop":
			ctx.flowManager.ControlFlow("STOP", id)

		}

		return c.NoContent(http.StatusOK)
	})

	ctx.echo.DELETE("/fimp/flow/definition/:id", func(c echo.Context) error {
		id := c.Param("id")
		ctx.flowManager.DeleteFlowFromStorage(id)
		return c.NoContent(http.StatusOK)
	})

}

func (ctx *FlowApi) RegisterMqttApi() {

}

type ImportFlowFromUrlRequest struct {
	Url   string
	Token string
}
