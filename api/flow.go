package api

import (
	"github.com/alivinco/tpflow/flow"
	"github.com/alivinco/tpflow/model"
	"github.com/labstack/echo"
	"io/ioutil"
	"net/http"
)

type FlowApi struct {
	flowManager *flow.Manager
	echo *echo.Echo
}

func NewFlowApi(flowManager *flow.Manager,echo *echo.Echo) *FlowApi {
	ctxApi := FlowApi{flowManager:flowManager,echo:echo}
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

	ctx.echo.POST("/fimp/flow/definition/:id", func(c echo.Context) error {
		id := c.Param("id")
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		ctx.flowManager.UpdateFlowFromJsonAndSaveToStorage(id, body)
		return c.NoContent(http.StatusOK)
	})

	ctx.echo.POST("/fimp/flow/definition/import", func(c echo.Context) error {
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		ctx.flowManager.ImportFlow(body)
		return c.NoContent(http.StatusOK)
	})


	ctx.echo.POST("/fimp/flow/ctrl/:id/:op", func(c echo.Context) error {
		id := c.Param("id")
		op := c.Param("op")

		switch op {
		case "send-inclusion-report" :
			ctx.flowManager.GetFlowById(id).SendInclusionReport()
		case "send-exclusion-report" :
			ctx.flowManager.GetFlowById(id).SendExclusionReport()
		case "start" :
			ctx.flowManager.ControlFlow("START",id)
		case "stop" :
			ctx.flowManager.ControlFlow("STOP",id)

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