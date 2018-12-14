package client

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/alivinco/fimpgo"
	conmodel "github.com/alivinco/tpflow/connector/model"
	"github.com/alivinco/tpflow/flow"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/registry"
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
	reqMsg := fimpgo.NewStrMapMessage("cmd.flow_ctx.get_records","tpflow",reqValue,nil,nil,nil)
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
	var reqValue map[string]interface{}
	reqValue["flow_id"] = flowId
	reqValue["rec"] = rec

	reqMsg := fimpgo.NewMessage("cmd.flow_ctx.update_record","tpflow","object",reqValue,nil,nil,nil)
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
	reqMsg := fimpgo.NewStrMapMessage("cmd.flow_ctx.delete","tpflow",cmdVal,nil,nil,nil)
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

func (rc *ApiRemoteClient) RegistryGetListOfServices() ([]registry.ThingWithLocationView,error) {
	return nil, nil

}







