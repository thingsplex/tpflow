package api

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/model"
	"github.com/labstack/echo"
	"io/ioutil"
	"net/http"
	"time"
)

type ContextApi struct {
	ctx  *model.Context
	echo *echo.Echo
	msgTransport *fimpgo.MqttTransport
}

func NewContextApi(ctx *model.Context, echo *echo.Echo) *ContextApi {
	ctxApi := ContextApi{ctx: ctx, echo: echo}
	ctxApi.RegisterRestApi()
	return &ctxApi
}

func (ctx *ContextApi) RegisterRestApi() {
	ctx.echo.GET("/fimp/api/flow/context/:flowid", func(c echo.Context) error {
		id := c.Param("flowid")
		if id != "-" {
			result := ctx.ctx.GetRecords(id)
			return c.JSON(http.StatusOK, result)
		}
		var result []model.ContextRecord
		return c.JSON(http.StatusOK, result)
	})

	ctx.echo.POST("/fimp/api/flow/context/record/:flowid", func(c echo.Context) error {
		flowId := c.Param("flowid")
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		rec := model.ContextRecord{}
		rec.UpdatedAt = time.Now()
		err = json.Unmarshal(body, &rec)
		if err != nil {
			log.Error("<ContextApi> Can't unmarshal context record.")
			return err
		}
		ctx.ctx.PutRecord(&rec, flowId, false)

		return c.JSON(http.StatusOK, rec)
	})

	ctx.echo.DELETE("/fimp/api/flow/context/record/:flowid", func(c echo.Context) error {
		// flowId is variable name here
		name := c.Param("flowid")
		log.Info("<ctx> Request to delete record with name ", name)
		if name != "" {
			err := ctx.ctx.DeleteRecord(name, "global", false)
			return c.JSON(http.StatusOK, err)
		}
		return c.JSON(http.StatusOK, nil)
	})
}

func (ctx *ContextApi) RegisterMqttApi(msgTransport *fimpgo.MqttTransport) {
	ctx.msgTransport = msgTransport
	ctx.msgTransport.Subscribe("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:1")
	apiCh := make(fimpgo.MessageCh, 10)
	ctx.msgTransport.RegisterChannel("flow-api",apiCh)
	var fimp *fimpgo.FimpMessage
	go func() {
		for {

			newMsg := <-apiCh
			log.Debug("New message of type ", newMsg.Payload.Type)
			switch newMsg.Payload.Type {
			case "cmd.flow.get_context":

				fimp = fimpgo.NewMessage("evt.flow.import_report", "tpflow", "string", resp, nil, nil, newMsg.Payload)

			}
			addr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeApp, ResourceName: "tpflow", ResourceAddress: "1",}
			ctx.msgTransport.Publish(&addr, fimp)
		}
	}()


}
