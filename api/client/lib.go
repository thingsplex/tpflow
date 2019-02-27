package client

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/futurehomeno/fimpgo"
	conmodel "github.com/alivinco/tpflow/connector/model"
	"github.com/alivinco/tpflow/flow"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/registry"
	"github.com/alivinco/tpflow/utils"
	"strconv"
)

type ApiRemoteClient struct {
	sClient * fimpgo.SyncClient
	timeout int64
	instanceAddress string
}

func NewApiRemoteClient(sClient *fimpgo.SyncClient, instanceAddress string) *ApiRemoteClient {
	sClient.AddSubscription("pt:j1/mt:evt/rt:app/rn:tpflow/ad:"+instanceAddress)
	return &ApiRemoteClient{sClient: sClient,instanceAddress:instanceAddress,timeout:15}
}

func (rc *ApiRemoteClient) GetListOfFlows()([]flow.FlowListItem,error) {
	reqMsg := fimpgo.NewNullMessage("cmd.flow.get_list","tpflow",nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return nil,err
	}

	var resp []flow.FlowListItem
	err = json.Unmarshal(respMsg.GetRawObjectValue(), &resp)
	if err != nil {
		log.Error("Can't unmarshal ", err)
		return nil , err
	}
	return resp,nil

}

func (rc *ApiRemoteClient) GetFlowDefinition(flowId string) (*model.FlowMeta,error) {
	reqMsg := fimpgo.NewStringMessage("cmd.flow.get_definition","tpflow",flowId,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return nil,err
	}
	var resp model.FlowMeta
	err = json.Unmarshal(respMsg.GetRawObjectValue(), &resp)
	if err != nil {
		log.Error("Can't unmarshal ", err)
		return nil,err
	}
	return &resp,nil

}

func (rc *ApiRemoteClient) GetConnectorTemplate(templateId string) (conmodel.Instance,error) {
	var resp conmodel.Instance
	reqMsg := fimpgo.NewStringMessage("cmd.flow.get_connector_template","tpflow",templateId,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return resp,err
	}
	err = json.Unmarshal(respMsg.GetRawObjectValue(), &resp)
	if err != nil {
		log.Error("Can't unmarshal ", err)
		return resp,err
	}
	return resp,nil
}

//cmd.flow.get_connector_template

func (rc *ApiRemoteClient) GetConnectorPlugins() (map[string]conmodel.Plugin,error) {
	var resp map[string]conmodel.Plugin
	reqMsg := fimpgo.NewNullMessage("cmd.flow.get_connector_plugins","tpflow",nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return resp,err
	}
	err = json.Unmarshal(respMsg.GetRawObjectValue(), &resp)
	if err != nil {
		log.Error("Can't unmarshal ", err)
		return resp,err
	}
	return resp,nil

}

func (rc *ApiRemoteClient) GetConnectorInstances() ([]conmodel.InstanceView,error) {
	var resp []conmodel.InstanceView
	reqMsg := fimpgo.NewNullMessage("cmd.flow.get_connector_instances","tpflow",nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return resp,err
	}
	err = json.Unmarshal(respMsg.GetRawObjectValue(), &resp)
	if err != nil {
		log.Error("Can't unmarshal ", err)
		return resp,err
	}
	return resp,nil

}

func (rc *ApiRemoteClient) ImportFlow(flowDef []byte) (string, error) {
	//var resp []conmodel.InstanceView
	var flowDefJson interface{}
	err := json.Unmarshal(flowDef,&flowDefJson)
	if err != nil {
		log.Error("Can't unmarshal ", err)
		return "",err
	}
	reqMsg := fimpgo.NewMessage("cmd.flow.import","tpflow","object",flowDefJson,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return "",err
	}
	if err != nil {
		log.Error("Can't unmarshal ", err)
		return "",err
	}
	return respMsg.GetStringValue()
}

func (rc *ApiRemoteClient) UpdateFlowBin(flowDef []byte) (string, error) {
	var flowDefJson interface{}
	err := json.Unmarshal(flowDef,&flowDefJson)
	if err != nil {
		log.Error("Can't unmarshal ", err)
		return "",err
	}
	reqMsg := fimpgo.NewMessage("cmd.flow.update_definition","tpflow","object",flowDefJson,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return "",err
	}

	return respMsg.GetStringValue()
}
// ControlFlow sends control command to flow manager.
// cmd - command to send , id - flow id
func (rc *ApiRemoteClient) ControlFlow(cmd string,id string) (string, error) {
	cmdVal := make(map[string]string)
	cmdVal["op"] = cmd
	cmdVal["id"] = id

	reqMsg := fimpgo.NewStrMapMessage("cmd.flow.ctrl","tpflow",cmdVal,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return "",err
	}
	return respMsg.GetStringValue()
}

func (rc *ApiRemoteClient) DeleteFlow(id string) (string, error) {
	reqMsg := fimpgo.NewStringMessage("cmd.flow.delete","tpflow",id,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return "",err
	}
	return respMsg.GetStringValue()
}

func (rc *ApiRemoteClient) ImportFlowFromUrl(url string, token string) (string, error) {
	cmdVal := make(map[string]string)
	cmdVal["url"] = url
	cmdVal["token"] = token

	reqMsg := fimpgo.NewStrMapMessage("cmd.flow.import_from_url","tpflow",cmdVal,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return "",err
	}
	return respMsg.GetStringValue()
}

func (rc *ApiRemoteClient) ContextGetRecords(flowId string) ([]model.ContextRecord, error) {
	var resp []model.ContextRecord
	reqValue := make(map[string]string)
	reqValue["flow_id"] = flowId
	reqMsg := fimpgo.NewStrMapMessage("cmd.flow.ctx_get_records","tpflow",reqValue,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return resp,err
	}
	err = json.Unmarshal(respMsg.GetRawObjectValue(), &resp)
	if err != nil {
		log.Error("Can't unmarshal ", err)
		return resp,err
	}
	return resp,nil
}

func (rc *ApiRemoteClient) ContextUpdateRecord(flowId string , rec *model.ContextRecord) (string,error) {
	reqValue := make(map[string]interface{})
	reqValue["flow_id"] = flowId
	reqValue["rec"] = rec

	reqMsg := fimpgo.NewMessage("cmd.flow.ctx_update_record","tpflow","object",reqValue,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return "",err
	}

	return respMsg.GetStringValue()
}

func (rc *ApiRemoteClient) ContextDeleteRecord(name string,flowId string) (string, error) {
	cmdVal := make(map[string]string)
	cmdVal["name"] = name
	cmdVal["flow_id"] = flowId
	reqMsg := fimpgo.NewStrMapMessage("cmd.flow.ctx_delete","tpflow",cmdVal,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return "",err
	}
	return respMsg.GetStringValue()
}

func (rc *ApiRemoteClient) RegistryGetListOfThings() ([]registry.ThingWithLocationView,error) {
	var resp []registry.ThingWithLocationView
	reqMsg := fimpgo.NewNullMessage("cmd.registry.get_things","tpflow",nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return resp,err
	}
	err = json.Unmarshal(respMsg.GetRawObjectValue(), &resp)
	if err != nil {
		log.Error("Can't unmarshal ", err)
		return resp,err
	}
	return resp,nil
}

func (rc *ApiRemoteClient) RegistryGetListOfServices(serviceName,locationId,thingId,filterWithoutAlias string) ([]registry.ServiceExtendedView,error) {
	var resp []registry.ServiceExtendedView
	cmdVal := make(map[string]string)
	cmdVal["service_name"] = serviceName
	cmdVal["location_id"] = locationId
	cmdVal["thing_id"] = thingId
	cmdVal["filter_without_alias"] = filterWithoutAlias

	reqMsg := fimpgo.NewStrMapMessage("cmd.registry.get_services","tpflow",cmdVal,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return nil,err
	}
	err = json.Unmarshal(respMsg.GetRawObjectValue(), &resp)
	if err != nil {
		log.Error("Can't unmarshal service response", err)
		return resp,err
	}
	return resp,nil
}

func (rc *ApiRemoteClient) RegistryGetListOfLocations() ([]registry.Location,error) {
	var resp []registry.Location
	cmdVal := make(map[string]string)
	reqMsg := fimpgo.NewStrMapMessage("cmd.registry.get_locations","tpflow",cmdVal,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return nil,err
	}
	err = json.Unmarshal(respMsg.GetRawObjectValue(), &resp)
	if err != nil {
		log.Error("Can't unmarshal service response", err)
		return resp,err
	}
	return resp,nil
}

func (rc *ApiRemoteClient) RegistryGetThing(tech,address string) (registry.ThingExtendedView,error) {
	var resp registry.ThingExtendedView
	cmdVal := make(map[string]string)
	cmdVal["tech"] = tech
	cmdVal["address"] = address

	reqMsg := fimpgo.NewStrMapMessage("cmd.registry.get_thing","tpflow",cmdVal,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return resp,err
	}
	err = json.Unmarshal(respMsg.GetRawObjectValue(), &resp)
	if err != nil {
		log.Error("Can't unmarshal thing response", err)
		return resp,err
	}
	return resp,nil
}

func (rc *ApiRemoteClient) RegistryGetService(fullAddress string) (registry.ServiceExtendedView,error) {
	var resp registry.ServiceExtendedView
	cmdVal := make(map[string]string)
	cmdVal["address"] =fullAddress

	reqMsg := fimpgo.NewStrMapMessage("cmd.registry.get_service","tpflow",cmdVal,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return resp,err
	}
	err = json.Unmarshal(respMsg.GetRawObjectValue(), &resp)
	if err != nil {
		log.Error("Can't unmarshal service response", err)
		return resp,err
	}
	return resp,nil
}

func (rc *ApiRemoteClient) RegistryUpdateLocationBin(locationBin []byte)(string, error) {
	var location registry.Location
	err := json.Unmarshal(locationBin,&location)
	if err != nil {
		log.Error("Can't unmarshal location ", err)
		return "",err
	}
	return rc.RegistryUpdateLocation(&location)
}

func (rc *ApiRemoteClient) RegistryUpdateLocation(location *registry.Location)(string, error) {
	reqMsg := fimpgo.NewMessage("cmd.registry.update_location","tpflow",fimpgo.VTypeObject,location,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return "",err
	}
	if err != nil {
		log.Error("Can't unmarshal ", err)
		return "",err
	}
	response,err := respMsg.GetStrMapValue()
	locationId , _ := response["id"]
	status , _ := response["status"]
	if status != "ok" {
		err = errors.New(status)
	}
	return locationId,err
}

func (rc *ApiRemoteClient) RegistryUpdateServiceBin(serviceBin []byte)(string, error) {
	var service registry.Service
	err := json.Unmarshal(serviceBin,&service)
	if err != nil {
		log.Error("Can't unmarshal service ", err)
		return "",err
	}
	return rc.RegistryUpdateService(&service)
}

func (rc *ApiRemoteClient) RegistryUpdateService(service *registry.Service)(string, error)  {
	reqMsg := fimpgo.NewMessage("cmd.registry.update_service","tpflow",fimpgo.VTypeObject,service,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return "",err
	}
	if err != nil {
		log.Error("Can't unmarshal ", err)
		return "",err
	}
	response,err := respMsg.GetStrMapValue()
	serviceId , _ := response["id"]
	status , _ := response["status"]
	if status != "ok" {
		err = errors.New(status)
	}
	return serviceId,err
}

func (rc *ApiRemoteClient) RegistryDeleteThing(id string) (string, error) {
	reqMsg := fimpgo.NewStringMessage("cmd.registry.delete_thing","tpflow",id,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return "",err
	}
	return respMsg.GetStringValue()
}

func (rc *ApiRemoteClient) RegistryDeleteLocation(id string) (string, error) {
	reqMsg := fimpgo.NewStringMessage("cmd.registry.delete_location","tpflow",id,nil,nil,nil)
	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return "",err
	}
	return respMsg.GetStringValue()
}

func (rc *ApiRemoteClient) GetFlowLog(limitLines int , flowId string)([]utils.LogEntry,error) {
	cmdVal := make(map[string]string)
	cmdVal["limit"] = strconv.Itoa(limitLines)
	cmdVal["flowId"] = flowId
	reqMsg := fimpgo.NewStrMapMessage("cmd.flow.get_log","tpflow",cmdVal,nil,nil,nil)

	respMsg , err := rc.sClient.SendFimp("pt:j1/mt:cmd/rt:app/rn:tpflow/ad:"+rc.instanceAddress,reqMsg,rc.timeout)
	if err != nil {
		return nil,err
	}

	var resp []utils.LogEntry
	err = json.Unmarshal(respMsg.GetRawObjectValue(), &resp)
	if err != nil {
		log.Error("Can't unmarshal log entries ", err)
		return nil , err
	}
	return resp,nil

}