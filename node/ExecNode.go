package node

import (
	"encoding/json"
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/model"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type ExecNode struct {
	BaseNode
	ctx *model.Context
	transport *fimpgo.MqttTransport
	config ExecNodeConfig
	scriptFullPath string
}

type ExecNodeConfig struct {
	ExecType string // cmd , sh-cmd , python , script
	Command string
	ScriptBody string
	InputVariableName string
	IsInputVariableGlobal bool
	OutputVariableName string
	IsOutputVariableGlobal bool
	IsOutputJson bool
	IsInputJson bool
}

func NewExecNode(flowOpCtx *model.FlowOperationalContext,meta model.MetaNode,ctx *model.Context,transport *fimpgo.MqttTransport) model.Node {
	node := ExecNode{ctx:ctx,transport:transport}
	node.meta = meta
	node.flowOpCtx = flowOpCtx
	node.config = ExecNodeConfig{}
	node.SetupBaseNode()
	return &node
}

func (node *ExecNode) LoadNodeConfig() error {
	err := mapstructure.Decode(node.meta.Config,&node.config)
	if err != nil{
		node.getLog().Error("Can't decode config file.Eee:",err)
	}
	if node.config.ExecType == "python" {
		node.scriptFullPath = filepath.Join(node.flowOpCtx.StoragePath,node.flowOpCtx.FlowId+"_"+string(node.meta.Id)+".py")
		err = ioutil.WriteFile(node.scriptFullPath, []byte(node.config.ScriptBody), 0644)
	}
	return err
}

// is invoked when node flow is stopped
func (node *ExecNode) Cleanup() error {
	if node.scriptFullPath != "" {
		os.Remove(node.scriptFullPath)
	}
	return nil
}

func (node *ExecNode) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *ExecNode) OnInput( msg *model.Message) ([]model.NodeID,error) {
	node.getLog().Info("Executing ExecNode . Name = ", node.meta.Label)

	//log.Debug(node.flowOpCtx.FlowId+"<ExecNode> Input value : ", r)
    var cmd * exec.Cmd
	switch node.config.ExecType {
	case "cmd":
		cmd = exec.Command(node.config.Command)
	case "sh-cmd":
		cmd = exec.Command("bash", "-c", node.config.Command)
	case "python":
		if node.config.IsInputJson {

			if msg.Payload.ValueType == "object" {
				var val interface{}
				msg.Payload.GetObjectValue(&val)
				msg.Payload.Value = val
				node.getLog().Debug("Input value : ", val)
			}
			strMsg,err := json.Marshal(msg)
			if err != nil {
				return []model.NodeID{node.meta.ErrorTransition},err
			}
			cmd = exec.Command("python3",node.scriptFullPath,string(strMsg))
		}else {
			cmd = exec.Command("python3",node.scriptFullPath)
		}
	}
	output , err := cmd.CombinedOutput()
	node.getLog().Debug("Normal output : ", string(output))
	if err != nil {
		node.getLog().Debug("Err output : ", err.Error())
	}


	flowId := node.flowOpCtx.FlowId
	outputJson := make(map[string]interface{})
	if node.config.IsOutputJson {
		err = json.Unmarshal(output,&outputJson)

	}
	if err != nil {
		node.getLog().Debug("Script output can't be parsed into JSON object: ", err.Error())
		return []model.NodeID{node.meta.ErrorTransition},err
	}

	if node.config.OutputVariableName != "" {
		if node.config.IsOutputVariableGlobal {
			flowId = "global"
		}
		if node.config.IsOutputJson {
			node.getLog().Debug("Output JSON : ", outputJson["ab"])
			err = node.ctx.SetVariable(node.config.OutputVariableName,"object",outputJson,"",flowId,false )
		}else {
			err = node.ctx.SetVariable(node.config.OutputVariableName,"string",string(output),"",flowId,false )
		}

	}else {
		if node.config.IsOutputJson {
			msg.Payload.Value = outputJson
			msg.Payload.ValueType = "object"
		}else {
			msg.Payload.Value = string(output)
			msg.Payload.ValueType = "string"
		}
	}

	if err != nil {
		node.getLog().Debug("Failed to save variable : ", err.Error())
		return []model.NodeID{node.meta.ErrorTransition},err
	}
	return []model.NodeID{node.meta.SuccessTransition},nil
}

