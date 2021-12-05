package api

import (
	"encoding/json"
	"fmt"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow"
	model2 "github.com/thingsplex/tpflow/connector/model"
	"github.com/thingsplex/tpflow/connector/plugins"
	httpcon "github.com/thingsplex/tpflow/connector/plugins/http"
	"github.com/thingsplex/tpflow/flow"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/utils"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

type FlowApi struct {
	flowManager *flow.Manager
	msgTransport *fimpgo.MqttTransport
	config       *tpflow.Configs
}

func NewFlowApi(flowManager *flow.Manager, config *tpflow.Configs) *FlowApi {
	ctxApi := FlowApi{flowManager: flowManager, config:config}
	return &ctxApi
}

//func (ctx *FlowApi) RegisterRestApi() {
//	ctx.echo.GET("/fimp/flow/list", func(c echo.Context) error {
//		resp := ctx.flowManager.GetFlowList()
//		return c.JSON(http.StatusOK, resp)
//	})
//	ctx.echo.GET("/fimp/flow/definition/:id", func(c echo.Context) error {
//		id := c.Param("id")
//		var resp *model.FlowMeta
//		if id == "-" {
//			flow := ctx.flowManager.GenerateNewFlow()
//			resp = &flow
//		} else {
//			resp = ctx.flowManager.GetFlowById(id).FlowMeta
//		}
//
//		return c.JSON(http.StatusOK, resp)
//	})
//
//	ctx.echo.GET("/fimp/connector/template/:id", func(c echo.Context) error {
//		id := c.Param("id")
//		result := plugins.GetConfigurationTemplate(id)
//		return c.JSON(http.StatusOK, result)
//	})
//
//	ctx.echo.GET("/fimp/connector/plugins", func(c echo.Context) error {
//		result := plugins.GetPlugins()
//		return c.JSON(http.StatusOK, result)
//	})
//
//	ctx.echo.GET("/fimp/connector/list", func(c echo.Context) error {
//		result := ctx.flowManager.GetConnectorRegistry().GetAllInstances()
//		return c.JSON(http.StatusOK, result)
//	})
//
//	ctx.echo.POST("/fimp/flow/definition/:id", func(c echo.Context) error {
//		id := c.Param("id")
//		body, err := ioutil.ReadAll(c.Request().Body)
//		if err != nil {
//			return err
//		}
//		ctx.flowManager.UpdateFlowFromBinJson(id, body)
//		return c.NoContent(http.StatusOK)
//	})
//
//	ctx.echo.PUT("/fimp/flow/definition/import", func(c echo.Context) error {
//		body, err := ioutil.ReadAll(c.Request().Body)
//		if err != nil {
//			return err
//		}
//		ctx.flowManager.ImportFlow(body)
//		return c.NoContent(http.StatusOK)
//	})
//
//	ctx.echo.PUT("/fimp/flow/definition/import_from_url", func(c echo.Context) error {
//
//		body, err := ioutil.ReadAll(c.Request().Body)
//		if err != nil {
//			return err
//		}
//		request := ImportFlowFromUrlRequest{}
//		err = json.Unmarshal(body, &request)
//		if err != nil {
//			log.Error("Can't parse request ", err)
//		}
//
//		// Get the data
//		resp, err := http.Get(request.Url)
//		if err != nil {
//			return err
//		}
//		flow, err := ioutil.ReadAll(resp.Body)
//		if err != nil {
//			log.Error("Can't read file from url ", err)
//			return err
//		}
//		log.Info("Importing flow")
//		ctx.flowManager.ImportFlow(flow)
//		return c.NoContent(http.StatusOK)
//	})
//
//	ctx.echo.POST("/fimp/flow/ctrl/:id/:op", func(c echo.Context) error {
//		id := c.Param("id")
//		op := c.Param("op")
//
//		switch op {
//		case "send-inclusion-report":
//			ctx.flowManager.GetFlowById(id).SendInclusionReport()
//		case "send-exclusion-report":
//			ctx.flowManager.GetFlowById(id).SendExclusionReport()
//		case "start":
//			ctx.flowManager.ControlFlow("START", id)
//		case "stop":
//			ctx.flowManager.ControlFlow("STOP", id)
//
//		}
//
//		return c.NoContent(http.StatusOK)
//	})
//
//	ctx.echo.DELETE("/fimp/flow/definition/:id", func(c echo.Context) error {
//		id := c.Param("id")
//		ctx.flowManager.DeleteFlowFromStorage(id)
//		return c.NoContent(http.StatusOK)
//	})
//
//}

func (ctx *FlowApi) RegisterMqttApi(msgTransport *fimpgo.MqttTransport,wsConn *httpcon.Connector) {
	ctx.msgTransport = msgTransport
	// TODO : Implement dynamic addressing and discovery
	ctx.msgTransport.Subscribe("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:1")
	ctx.msgTransport.Subscribe("pt:j1/mt:evt/rt:ad/rn:gateway/ad:1")
	ctx.msgTransport.Subscribe("pt:j1/mt:evt/rt:ad/+/+") // Adapter events for flow auto configuration based on added products

	apiCh := make(fimpgo.MessageCh, 10)
	ctx.msgTransport.RegisterChannel("flow-api",apiCh)
	var fimp *fimpgo.FimpMessage
	go func() {
		for {
			newMsg := <-apiCh
			fimp = nil
			log.Debug("New flow message of type ", newMsg.Payload.Type)
			switch newMsg.Payload.Type {
			case "cmd.flow.get_list":
				val := ctx.flowManager.GetFlowList()
				fimp = fimpgo.NewMessage("evt.flow.list_report", "tpflow", "object", val, nil, nil, newMsg.Payload)

			case "cmd.flow.get_definition":
				var resp *model.FlowMeta
				id, _ := newMsg.Payload.GetStringValue()
				if id == "-" {
					flow := ctx.flowManager.GenerateNewFlow()
					resp = &flow
				} else {
					resp = ctx.flowManager.GetFlowById(id).FlowMeta
				}
				fimp = fimpgo.NewMessage("evt.flow.definition_report", "tpflow", "object", resp, nil, nil, newMsg.Payload)

			case "cmd.flow.get_connector_template":
				id, _ := newMsg.Payload.GetStringValue()
				resp := plugins.GetConfigurationTemplate(id)
				fimp = fimpgo.NewMessage("cmd.flow.connector_template_report", "tpflow", "object", resp, nil, nil, newMsg.Payload)

			case "cmd.flow.get_connector_plugins":
				resp := plugins.GetPlugins()
				fimp = fimpgo.NewMessage("evt.flow.connector_plugins_report", "tpflow", "object", resp, nil, nil, newMsg.Payload)

			case "cmd.flow.get_connector_instances":
				resp := ctx.flowManager.GetConnectorRegistry().GetAllInstances()
				fimp = fimpgo.NewMessage("evt.flow.connector_instances_report", "tpflow", "object", resp, nil, nil, newMsg.Payload)

			case "cmd.flow.update_connector_instance_config":
				resp := "ok"
				instConf := model2.Instance{}
				err := newMsg.Payload.GetObjectValue(&instConf)
				if err != nil {
					log.Error("<ConnRegistry> Incorrect payload for cmd.flow.update_connector_instance.")
					resp = err.Error()
				}else {
					err = ctx.flowManager.GetConnectorRegistry().UpdateConnectorConfig(instConf.ID,instConf.Config)
					if err != nil {
						resp = err.Error()
					}
				}

				fimp = fimpgo.NewMessage("evt.flow.connector_instances_update_report", "tpflow", "string", resp, nil, nil, newMsg.Payload)

			case "cmd.flow.update_definition":
				flowMeta := model.FlowMeta{}
				flowJsonDef := newMsg.Payload.GetRawObjectValue()
				err := json.Unmarshal(flowJsonDef, &flowMeta)
				if err != nil {
					log.Error("<FlMan> Can't unmarshel flow definition.")
					fimp = fimpgo.NewMessage("evt.flow.update_report", "tpflow", "string", err, nil, nil, newMsg.Payload)
					break
				}
				err = ctx.flowManager.UpdateFlowFromBinJson(flowMeta.Id, flowJsonDef)
				report := "ok"
				if err != nil {
					report = err.Error()
				}
				fimp = fimpgo.NewMessage("evt.flow.update_report", "tpflow", "string", report, nil, nil, newMsg.Payload)

			case "cmd.flow.import":
				resp := "ok"
				err := ctx.flowManager.ImportFlow(newMsg.Payload.GetRawObjectValue())
				if err != nil {
					resp = err.Error()
				}
				fimp = fimpgo.NewMessage("evt.flow.import_report", "tpflow", "string", resp, nil, nil, newMsg.Payload)
			case "cmd.backup.execute":
				err := ctx.flowManager.BackupAll()
				op := "ok"
				errStr := ""
				if err != nil {
					op = "error"
					errStr = err.Error()
				}
				resp := map[string]string {"op_status":op,"error":errStr}
				fimp = fimpgo.NewStrMapMessage("evt.backup.report", "tpflow", resp, nil, nil, newMsg.Payload)

			case "cmd.flow.ctrl":
				resp := "ok"

				val, err := newMsg.Payload.GetStrMapValue()
				if err != nil {
					log.Error("Wrong value format ")
					fimp = fimpgo.NewMessage("evt.flow.ctrl_report", "tpflow", "string", err.Error(), nil, nil, newMsg.Payload)
					break

				}
				op, ok1 := val["op"]
				id, ok2 := val["id"]

				if !ok1 || !ok2 {
					fimp = fimpgo.NewMessage("evt.flow.ctrl_report", "tpflow", "string", "missing param", nil, nil, newMsg.Payload)
					break
				}
				switch op {
				case "send-inclusion-report":
					ctx.flowManager.GetFlowById(id).SendInclusionReport()
				case "send-exclusion-report":
					ctx.flowManager.GetFlowById(id).SendExclusionReport()
				case "start":
					err = ctx.flowManager.ControlFlow("START", id)
				case "stop":
					err = ctx.flowManager.ControlFlow("STOP", id)

				}
				if err != nil {
					resp = err.Error()
				}
				fimp = fimpgo.NewMessage("evt.flow.ctr_report", "tpflow", "string", resp, nil, nil, newMsg.Payload)

			case "cmd.flow.delete":
				resp := "ok"
				id,err := newMsg.Payload.GetStringValue()
				if err == nil {
					ctx.flowManager.DeleteFlowFromStorage(id)
				}else {
					resp = err.Error()
				}
				fimp = fimpgo.NewMessage("evt.flow.delete_report", "tpflow", "string", resp, nil, nil, newMsg.Payload)

			case "cmd.flow.import_from_url":
				resp := "ok"
				val, err := newMsg.Payload.GetStrMapValue()
				if err != nil {
					log.Error("Wrong value format ")
					fimp = fimpgo.NewMessage("evt.flow.import_report", "tpflow", "string", err.Error(), nil, nil, newMsg.Payload)
					break

				}
				url, ok := val["url"]
				if !ok {
					log.Error("Url is not defined ")
					fimp = fimpgo.NewMessage("evt.flow.import_report", "tpflow", "string", err.Error(), nil, nil, newMsg.Payload)
					break
				}
				// Get the data
				hresponse, err := http.Get(url)
				//token ,ok := val["token"]
				if err != nil {
					log.Error("Can't load file from url , error = ", err)
					fimp = fimpgo.NewMessage("evt.flow.import_report", "tpflow", "string", err.Error(), nil, nil, newMsg.Payload)
					break
				}
				bflow, err := ioutil.ReadAll(hresponse.Body)
				if err != nil {
					log.Error("Can't read file from url ", err)
					fimp = fimpgo.NewMessage("evt.flow.import_report", "tpflow", "string", err.Error(), nil, nil, newMsg.Payload)
					break
				}
				log.Info("Importing flow")
				resp = "ok"
				if err := ctx.flowManager.ImportFlow(bflow); err != nil {
					resp = err.Error()
				}
				fimp = fimpgo.NewMessage("evt.flow.import_report", "tpflow", "string", resp, nil, nil, newMsg.Payload)
			case "cmd.flow.create_from_template":
				val, err := newMsg.Payload.GetStrMapValue()
				if err != nil {
					log.Error("Can't get log , wrong params , error = ", err)
					break
				}
				flowName , ok := val["flow_name"]
				if !ok {
					break
				}
				ac := flow.NewAutoConfig(ctx.config.FlowStorageDir)
				err = ac.LoadFlowFromTemplate(flowName)
				if err != nil {
					log.Error("<api> Can't load flow from template ",err)
					break
				}
				settings := map[string]model.Setting{}

				for k,v := range val {
					settings[k] = model.Setting{Value: v,ValueType: "string"}
				}
				ac.SetSettings(settings)
				ac.SaveNewFlow()
				err = ctx.flowManager.LoadFlowFromFile(ctx.flowManager.GetFlowFileNameById(flowName))
				if err != nil {
					log.Error("<api>Flow can't load . Error:",err)
				}else {
					log.Info("<api> Flow loaded successfully")
				}

			case "cmd.flow.get_log":
				val, err := newMsg.Payload.GetStrMapValue()
				if err != nil {
					log.Error("Can't get log , wrong params , error = ", err)
					break
				}
				flowId , _ := val["flowId"]
				limitS  , _ := val["limit"]
				limit , err := strconv.Atoi(limitS)
				if err != nil {
					limit = 10000
				}
				filter := utils.LogFilter{FlowId:flowId}
				log.Debug("Getting log from file :",ctx.config.LogFile)
				result := utils.GetLogs(ctx.config.LogFile,&filter,limit)
				fimp = fimpgo.NewMessage("evt.flow.log_report", "tpflow", "object", result, nil, nil, newMsg.Payload)

			case "cmd.flow.run_gc":
				log.Info("Running GC")
				runtime.GC()

			case "cmd.log.set_level":
				level , err := newMsg.Payload.GetStringValue()
				if err != nil {
					log.Error("<api> wrong payload type")
				}
				logLevel, err := log.ParseLevel(level)
				if err == nil {
					log.SetLevel(logLevel)
					//mex.configs.LogLevel = level
					//mex.configs.Save()
					log.Info("<msgex> Log level was updated to = ",level)
				}else {
					log.Error("<msgex> Unsupported log level = ",level)
				}
			case "evt.gateway.factory_reset","cmd.flow.factory_reset":
				if newMsg.Payload.Service == "gateway" || newMsg.Payload.Service == "tpflow" {
					log.Info("----- FACTORY RESET COMMAND -------------------")
					ctx.flowManager.FactoryReset()
					time.Sleep(1 * time.Second)
					os.Exit(1)
				}else {
					log.Error("<api> Cmd evt.gateway.factory_reset must have service gateway. ")
				}
			case "evt.thing.inclusion_report":
				// do product specific flow configuration here
				// Get product address , tech and hash
				// Load flow which starts with prod_auto_flow_zw_398_4_2.json
				// Bind flow with device - load flow -> set prod hash , device id and tech into settings -> save flow to flow folder
				log.Info("<api> Loading flow from template")
				inclReport := &fimptype.ThingInclusionReport{}
				err := newMsg.Payload.GetObjectValue(inclReport)
				if err != nil {
					break
				}
				settings := map[string]model.Setting{}
				settings["dev.address"] = model.Setting{Value:inclReport.Address,ValueType: "string",Description: "device address"}
				settings["dev.tech"] = model.Setting{Value:inclReport.CommTechnology,ValueType: "string",Description: "communication technology"}
				prodHash := inclReport.ProductHash
				settings["dev.prod.hash"] = model.Setting{Value:prodHash,ValueType: "string",Description: "hash code of the product"}
				if ctx.flowManager.GetFlowBySettings(settings) != nil {
					log.Info("<api> Flow with similar setting already registered . Operation skipped")
					break
				}
				flowName := fmt.Sprintf("auto_config_prod_hash_%s",prodHash)
				ac := flow.NewAutoConfig(ctx.config.FlowStorageDir)
				err = ac.LoadFlowFromTemplate(flowName)
				if err != nil {
					log.Debug("<api> Can't load flow from template . Err: ",err.Error())
					break
				}
				ac.SetSettings(settings)
				ac.SaveNewFlow()
				err = ctx.flowManager.LoadFlowFromFile(ctx.flowManager.GetFlowFileNameById(ac.Flow().Id))
				if err != nil {
					log.Error("<api>Flow can't be loaded . Error:",err)
				}else {
					log.Info("<api> Flow loaded successfully")
					ctx.flowManager.StartFlow(ac.Flow().Id)
				}

			case "evt.thing.exclusion_report":
				exclReport := &fimptype.ThingExclusionReport{}
				err := newMsg.Payload.GetObjectValue(exclReport)
				if err != nil {
					break
				}
				tech := newMsg.Addr.ResourceName
				addr := exclReport.Address
				log.Infof("<api> Deleting auto flow , tech = %s , addr = %s ",tech,addr)
				settings := map[string]model.Setting{}
				settings["dev.address"] = model.Setting{Value:addr,ValueType: "string",Description: "device address"}
				settings["dev.tech"] = model.Setting{Value:tech,ValueType: "string",Description: "communication technology"}
				flow := ctx.flowManager.GetFlowBySettings(settings)
				if flow != nil {
					if !flow.FlowMeta.IsDefault {
						log.Infof("<api> Deleting flow , id = %s ",flow.Id)
						ctx.flowManager.DeleteFlowFromStorage(flow.Id)
					}
				}else {
					log.Debug("<api> Flow not found ")
				}

			}

			if fimp != nil {
				addr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeApp, ResourceName: "tpflow", ResourceAddress: "1",}
				if err := ctx.msgTransport.RespondToRequest(newMsg.Payload,fimp); err !=nil {
					ctx.msgTransport.Publish(&addr, fimp)
				}
				if wsConn != nil {
					var reqId int64
					binFimp,err := fimp.SerializeToJson()
					if err != nil {
						// Sending response to the same connection we received request from
						wsConn.PublishToConnectionById(reqId,binFimp)
					}

				}
			}
		}
	}()


}

type ImportFlowFromUrlRequest struct {
	Url   string
	Token string
}
