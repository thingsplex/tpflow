package exec

import (
	"encoding/json"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/node/base"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type Node struct {
	base.BaseNode
	ctx            *model.Context
	config         NodeConfig
	scriptFullPath string
}

type NodeConfig struct {
	ExecType               string // cmd , sh-cmd , python , script
	Command                string
	ScriptBody             string
	InputVariableName      string
	IsInputVariableGlobal  bool
	OutputVariableName     string
	IsOutputVariableGlobal bool
	IsOutputJson           bool
	IsInputJson            bool
}

func NewNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := Node{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	node.config = NodeConfig{}
	node.SetupBaseNode()
	return &node
}

func (node *Node) LoadNodeConfig() error {
	err := mapstructure.Decode(node.Meta().Config, &node.config)
	if err != nil {
		node.GetLog().Error("Can't decode config file.Eee:", err)
	}
	if node.config.ExecType == "python" {
		node.scriptFullPath = filepath.Join(node.FlowOpCtx().StoragePath, node.FlowOpCtx().FlowId+"_"+string(node.Meta().Id)+".py")
		err = ioutil.WriteFile(node.scriptFullPath, []byte(node.config.ScriptBody), 0644)
	}
	return err
}

// is invoked when node flow is stopped
func (node *Node) Cleanup() error {
	if node.scriptFullPath != "" {
		os.Remove(node.scriptFullPath)
	}
	return nil
}

func (node *Node) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Info("Executing Node . Name = ", node.Meta().Label)

	//log.Debug(node.FlowOpCtx().FlowId+"<Node> Input value : ", r)
	var cmd *exec.Cmd
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
				node.GetLog().Debug("Input value : ", val)
			}
			strMsg, err := json.Marshal(msg)
			if err != nil {
				return []model.NodeID{node.Meta().ErrorTransition}, err
			}
			cmd = exec.Command("python3", node.scriptFullPath, string(strMsg))
		} else {
			cmd = exec.Command("python3", node.scriptFullPath)
		}
	}
	output, err := cmd.CombinedOutput()
	node.GetLog().Debug("Normal output : ", string(output))
	if err != nil {
		node.GetLog().Debug("Err output : ", err.Error())
	}

	flowId := node.FlowOpCtx().FlowId
	outputJson := make(map[string]interface{})
	if node.config.IsOutputJson {
		err = json.Unmarshal(output, &outputJson)

	}
	if err != nil {
		node.GetLog().Debug("Script output can't be parsed into JSON object: ", err.Error())
		return []model.NodeID{node.Meta().ErrorTransition}, err
	}

	if node.config.OutputVariableName != "" {
		if node.config.IsOutputVariableGlobal {
			flowId = "global"
		}
		if node.config.IsOutputJson {
			node.GetLog().Debug("Output JSON : ", outputJson)
			err = node.ctx.SetVariable(node.config.OutputVariableName, "object", outputJson, "", flowId, false)
		} else {
			err = node.ctx.SetVariable(node.config.OutputVariableName, "string", string(output), "", flowId, false)
		}

	} else {
		if node.config.IsOutputJson {
			msg.Payload.Value = outputJson
			msg.Payload.ValueType = "object"
		} else {
			msg.Payload.Value = string(output)
			msg.Payload.ValueType = "string"
		}
	}

	if err != nil {
		node.GetLog().Debug("Failed to save variable : ", err.Error())
		return []model.NodeID{node.Meta().ErrorTransition}, err
	}
	return []model.NodeID{node.Meta().SuccessTransition}, nil
}
