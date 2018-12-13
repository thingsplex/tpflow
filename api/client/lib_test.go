package client

import (
	"encoding/json"
	"github.com/alivinco/fimpgo"
	"testing"
)

func getApiRemoteClient()*ApiRemoteClient  {
	sClient := fimpgo.NewSyncClient(nil)
	sClient.Connect("tcp://localhost:1883","api_client_test","","",true,1,1)

	remoteApiClient := NewApiRemoteClient(sClient, "1")
	return remoteApiClient
}

func TestApiRemoteClient_GetListOfFlows(t *testing.T) {

	remoteApiClient := getApiRemoteClient()
	resp,err := remoteApiClient.GetListOfFlows()

	if err != nil || len(resp)==0 {
		t.Error(err)
	}else {
		t.Log(resp)
	}


}

func TestApiRemoteClient_GetFlowDefinition(t *testing.T) {
	remoteApiClient := getApiRemoteClient()
	resp,err := remoteApiClient.GetFlowDefinition("2t5DPqKrob48sRK")
	if err != nil {
		t.Error(err)
	}else {
		t.Log(resp)
	}
}


func TestApiRemoteClient_GetConnectorTemplate(t *testing.T) {
	remoteApiClient := getApiRemoteClient()
	resp,err := remoteApiClient.GetConnectorTemplate("fimpmqtt")
	if err != nil {
		t.Error(err)
	}else {
		t.Log(resp)
	}
}

func TestApiRemoteClient_GetConnectorPlugins(t *testing.T) {
	remoteApiClient := getApiRemoteClient()
	resp,err := remoteApiClient.GetConnectorPlugins()
	if err != nil {
		t.Error(err)
	}else {
		t.Log(resp)
	}
}

func TestApiRemoteClient_GetConnectorInstances(t *testing.T) {
	remoteApiClient := getApiRemoteClient()
	resp,err := remoteApiClient.GetConnectorInstances()
	if err != nil {
		t.Error(err)
	}else {
		t.Log(resp)
	}
}

func TestApiRemoteClient_UpdateFlowBin(t *testing.T) {
	remoteApiClient := getApiRemoteClient()

	resp,err := remoteApiClient.GetFlowDefinition("-")
	if err != nil {
		t.Error(err)
	}else {
		t.Log(resp)
	}

	resp.Name = "Test flow"

	flowMetaBin,err := json.Marshal(resp)

	updateResp,err := remoteApiClient.UpdateFlowBin(flowMetaBin)
	if err == nil && updateResp == "ok"  {
		t.Log(updateResp)
	}else {
		t.Error(err)
	}
}

func TestApiRemoteClient_ControlFlow(t *testing.T) {
	remoteApiClient := getApiRemoteClient()
	updateResp,err := remoteApiClient.ControlFlow("stop","Mw9Vs2DEEhwg4WB")
	if err == nil && updateResp == "ok"  {
		t.Log(updateResp)
	}else {
		t.Error(err)
	}
}


func TestApiRemoteClient_RegistryGetListOfThings(t *testing.T) {
	remoteApiClient := getApiRemoteClient()
	resp,err := remoteApiClient.RegistryGetListOfThings()

	if err != nil || len(resp)==0 {
		t.Error(err)
	}else {
		t.Log(resp)
	}
}


