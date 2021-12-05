package api

import (
	"encoding/json"
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/flow/context"
	"time"
)

type ContextApi struct {
	ctx  *context.Context
	msgTransport *fimpgo.MqttTransport
}

func NewContextApi(ctx *context.Context) *ContextApi {
	ctxApi := ContextApi{ctx: ctx}
	//ctxApi.RegisterRestApi()
	return &ctxApi
}

//func (ctx *ContextApi) RegisterRestApi() {
//	ctx.echo.GET("/fimp/api/flow/context/:flowid", func(c echo.Context) error {
//		id := c.Param("flowid")
//		if id != "-" {
//			result := ctx.ctx.GetRecords(id)
//			return c.JSON(http.StatusOK, result)
//		}
//		var result []model.ContextRecord
//		return c.JSON(http.StatusOK, result)
//	})
//
//	ctx.echo.POST("/fimp/api/flow/context/record/:flowid", func(c echo.Context) error {
//		flowId := c.Param("flowid")
//		body, err := ioutil.ReadAll(c.Request().Body)
//		if err != nil {
//			return err
//		}
//		rec := model.ContextRecord{}
//		rec.UpdatedAt = time.Now()
//		err = json.Unmarshal(body, &rec)
//		if err != nil {
//			log.Error("<ContextApi> Can't unmarshal context record.")
//			return err
//		}
//		ctx.ctx.PutRecord(&rec, flowId, false)
//
//		return c.JSON(http.StatusOK, rec)
//	})
//
//	ctx.echo.DELETE("/fimp/api/flow/context/record/:flowid", func(c echo.Context) error {
//		// flowId is variable name here
//		name := c.Param("flowid")
//		log.Info("<ctx> Request to delete record with name ", name)
//		if name != "" {
//			err := ctx.ctx.DeleteRecord(name, "global", false)
//			return c.JSON(http.StatusOK, err)
//		}
//		return c.JSON(http.StatusOK, nil)
//	})
//}

type ContextExtRecord struct {
	FlowId string                `json:"flow_id"`
	Rec    context.ContextRecord `json:"rec"`
}

func (ctx *ContextApi) RegisterMqttApi(msgTransport *fimpgo.MqttTransport) {
	ctx.msgTransport = msgTransport
	ctx.msgTransport.Subscribe("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:1")
	apiCh := make(fimpgo.MessageCh, 10)
	ctx.msgTransport.RegisterChannel("flow-ctx-api",apiCh)
	// TODO : register channel in WS transport , so that WS can generates messages along with request-id
	var fimp *fimpgo.FimpMessage
	go func() {
		for {
			newMsg := <-apiCh
			fimp = nil
			log.Debug("New context message of type ", newMsg.Payload.Type)
			switch newMsg.Payload.Type {
			case "cmd.flow.ctx_get_records":
				val,_ := newMsg.Payload.GetStrMapValue()
				flowId , _ := val["flow_id"]
				if flowId == "-"|| flowId=="" {
					flowId = "global"
				}
				result := ctx.ctx.GetRecords(flowId)
				fimp = fimpgo.NewMessage("evt.flow.ctx_records_report", "tpflow", fimpgo.VTypeObject, result, nil, nil, newMsg.Payload)

			case "cmd.flow.ctx_update_record":
				var reqValue ContextExtRecord
				reqRawObject := newMsg.Payload.GetRawObjectValue()
				err := json.Unmarshal(reqRawObject, &reqValue)
				if err != nil {
					log.Error("<ctx> cmd.flow.ctx_update_record Can't unmarshal request")
					fimp = fimpgo.NewMessage("evt.flow.ctx_update_report", "tpflow", "string", err.Error(), nil, nil, newMsg.Payload)
					break
				}
				reqValue.Rec.UpdatedAt = time.Now()
				ctx.ctx.PutRecord(&reqValue.Rec, reqValue.FlowId, reqValue.Rec.InMemory)
				fimp = fimpgo.NewMessage("evt.flow.ctx_update_report", "tpflow", "string", "ok", nil, nil, newMsg.Payload)

			case "cmd.flow.ctx_delete":
				// flowId is variable name here
				req ,err := newMsg.Payload.GetStrMapValue()
				if err != nil {
					log.Error("<ctx> Can't unmarshal request.")
					fimp = fimpgo.NewMessage("evt.flow.ctx_delete_report", "tpflow", "string", err, nil, nil, newMsg.Payload)
	                break
				}
				log.Info("<ctx> Request to delete record with name ", req["name"])
				flowId := "global"
				if req["flow_id"] != "" {
					flowId = req["flow_id"]
				}
				err = ctx.ctx.DeleteRecord(req["name"],flowId , false)
				if err != nil {
					fimp = fimpgo.NewMessage("evt.flow.ctx_delete_report", "tpflow", "string", err, nil, nil, newMsg.Payload)

				}else {
					fimp = fimpgo.NewMessage("evt.flow.ctx_delete_report", "tpflow", "string", "ok", nil, nil, newMsg.Payload)

				}
			}

			if fimp != nil {
				addr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeApp, ResourceName: "tpflow", ResourceAddress: "1",}
				ctx.msgTransport.Publish(&addr, fimp)
				// TODO : For WS publisher , the response must be sent to the same connection ,
			}

		}
	}()


}
