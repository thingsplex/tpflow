package rest

import (
	"bytes"
	"encoding/json"
	"github.com/ChrisTrenkamp/goxpath"
	"github.com/ChrisTrenkamp/goxpath/tree"
	"github.com/ChrisTrenkamp/goxpath/tree/xmltree"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"github.com/mitchellh/mapstructure"
	"github.com/oliveagle/jsonpath"
	"net/http"
	"text/template"

	"errors"
	"io/ioutil"
	"net/url"
	"strings"
	"time"
)

// Node
type Node struct {
	base.BaseNode
	ctx            *model.Context
	config         NodeConfig
	reqTemplate    *template.Template
	urlTemplate    *template.Template
	httpClient     *http.Client
	accessToken    string
	tokenExpiresAt time.Time
}

type ResponseToVariableMap struct {
	Name                 string
	Path                 string
	PathType             string // xml , json
	TargetVariableName   string
	IsVariableGlobal     bool
	TargetVariableType   string
	UpdateTriggerMessage bool
}

type Header struct {
	Name  string
	Value string
}

type NodeConfig struct {
	Url                  string
	Method               string // GET,POST,PUT,DELETE, etc.
	TemplateVariableName string
	IsVariableGlobal     bool
	RequestPayloadType   string // json,xml,string
	RequestTemplate      string
	Headers              []Header
	HeadersVariableName  string // header variable should be of type map_str
	ResponseMapping      []ResponseToVariableMap
	LogResponse          bool
	Auth                 OAuth
}

type OAuth struct {
	Enabled      bool
	GrantType    string
	Url          string
	ClientID     string
	ClientSecret string
	Scope        string
	Username     string
	Password     string
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RestActionNodeTemplateParams struct {
	Variable interface{}
	Token    string
	Message  *model.Message
}

func NewNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := Node{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	node.config = NodeConfig{}
	node.httpClient = &http.Client{Timeout:60*time.Second}
	node.SetupBaseNode()
	return &node
}

func (node *Node) LoadNodeConfig() error {
	err := mapstructure.Decode(node.Meta().Config, &node.config)

	funcMap := template.FuncMap{
		"variable": func(varName string, isGlobal bool) (interface{}, error) {
			//node.GetLog().Debug("Getting variable by name ",varName)
			var vari model.Variable
			var err error
			if isGlobal {
				vari, err = node.ctx.GetVariable(varName, "global")
			} else {
				vari, err = node.ctx.GetVariable(varName, node.FlowOpCtx().FlowId)
			}

			if vari.IsNumber() {
				return vari.ToNumber()
			}
			vstr, ok := vari.Value.(string)
			if ok {
				return vstr, err
			} else {
				return "", errors.New("Only simple types are supported ")
			}
		},
		"setting": func(name string) (interface{}, error) {
			if node.FlowOpCtx().FlowMeta.Settings != nil {
				s := node.FlowOpCtx().FlowMeta.Settings[name]
				return s.String(),nil
			}else {
				return "",nil
			}
		},
	}

	if err != nil {
		node.GetLog().Error(" Failed while loading configurations.Error:", err)

	} else {
		node.reqTemplate, err = template.New("request").Funcs(funcMap).Parse(node.config.RequestTemplate)
		if err != nil {
			node.GetLog().Error(" Failed while parsing request template.Error:", err)
		}
		node.urlTemplate, err = template.New("url").Funcs(funcMap).Parse(node.config.Url)
		if err != nil {
			node.GetLog().Error(" Failed while parsing url template.Error:", err)
		}
		node.Authenticate("", "")
	}
	return err
}

func (node *Node) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *Node) loadAccessTokenFromContext() bool {
	accessTokenVar, err := node.ctx.GetVariable("access_token", node.FlowOpCtx().FlowId)

	if err != nil {
		node.GetLog().Info(" Error , can't load access token from context.Err", err)
		node.accessToken = ""
		return false
	}
	var ok bool
	node.accessToken, ok = accessTokenVar.Value.(string)
	if !ok {
		node.GetLog().Info(" Error , can't load access token from context,variable is not a string .Err", err)
		node.accessToken = ""
		return false
	}

	if node.accessToken == "" {
		return false
	}
	return true
}

func (node *Node) Authenticate(username string, password string) {
	if !node.config.Auth.Enabled {
		node.accessToken = ""
		return
	}
	if node.loadAccessTokenFromContext() {
		node.GetLog().Info(" Auth token already loaded ")
		return
	}
	node.GetLog().Info(" Requesting new token")
	data := url.Values{}
	data.Add("grant_type", node.config.Auth.GrantType)
	data.Add("client_id", node.config.Auth.ClientID)
	data.Add("client_secret", node.config.Auth.ClientSecret)
	data.Add("scope", node.config.Auth.Scope)
	if node.config.Auth.GrantType == "authorization_code" {
		data.Add("code", node.config.Auth.Password)
	}
	if node.config.Auth.GrantType == "password" {
		data.Add("username", node.config.Auth.Username)
		data.Add("password", node.config.Auth.Password)
	}

	client := &http.Client{}
	r, _ := http.NewRequest("POST", node.config.Auth.Url, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	//bReq ,err := ioutil.ReadAll(r.Body)
	//node.GetLog().Info("<Node> Auth request:",string(bReq))

	resp, err := client.Do(r)
	if err != nil {
		node.GetLog().Info(" Auth error . Error : ", err)
		return
	}
	bData, err := ioutil.ReadAll(resp.Body)
	if err == nil {

		node.GetLog().Info("Auth Response:", string(bData))

		jresp := AuthResponse{}
		json.Unmarshal([]byte(bData), &jresp)

		node.GetLog().Info("Auth Response:", string(bData))

		node.accessToken = jresp.AccessToken
		node.tokenExpiresAt = time.Now().Add(time.Second * time.Duration(jresp.ExpiresIn))

		node.GetLog().Infof(" Auth successfully . Access toke expires after %d seconds , at %s", jresp.ExpiresIn, node.tokenExpiresAt.Format("2006-01-02 15:04:05"))
		node.ctx.SetVariable("access_token", "object", jresp.AccessToken, "", node.FlowOpCtx().FlowId, false)
		node.ctx.SetVariable("refresh_token", "string", jresp.RefreshToken, "", node.FlowOpCtx().FlowId, false)
		node.ctx.SetVariable("expires_at", "string", node.tokenExpiresAt, "", node.FlowOpCtx().FlowId, false)
	}

}

func (node *Node) refreshToken() {
	node.GetLog().Info(" Refreshing access token . ")
	refreshTokenVar, err := node.ctx.GetVariable("refresh_token", node.FlowOpCtx().FlowId)

	if err != nil {
		node.GetLog().Info(" Error , can't load refresh token from context.Err", err)
		node.accessToken = ""
		return
	}
	if refreshTokenVar.Value == "" {
		node.GetLog().Debug(" Refresh token is empty ")
		return
	}
	refreshToken, ok := refreshTokenVar.Value.(string)
	if !ok {
		node.GetLog().Info(" Error , can't load refresh token from context,variable is not a string .Err", err)
		node.accessToken = ""
		return
	}
	data := url.Values{}
	data.Add("grant_type", "refresh_token")
	data.Add("client_id", node.config.Auth.ClientID)
	data.Add("client_secret", node.config.Auth.ClientSecret)
	data.Add("refresh_token", refreshToken)

	client := &http.Client{}
	r, _ := http.NewRequest("POST", node.config.Auth.Url, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(r)
	if err != nil {
		node.GetLog().Info(" Auth error . Error : ", err)
		return
	}
	bData, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		jresp := AuthResponse{}
		json.Unmarshal([]byte(bData), &jresp)

		node.GetLog().Info("Auth Response:", string(bData))

		node.accessToken = jresp.AccessToken
		node.tokenExpiresAt = time.Now().Add(time.Second * time.Duration(jresp.ExpiresIn))

		node.GetLog().Infof(" Token refreshed successfully . Access toke expires after %d seconds , at %s", jresp.ExpiresIn, node.tokenExpiresAt.Format("2006-01-02 15:04:05"))
		node.ctx.SetVariable("access_token", "string", jresp.AccessToken, "", node.FlowOpCtx().FlowId, false)
		node.ctx.SetVariable("refresh_token", "string", jresp.RefreshToken, "", node.FlowOpCtx().FlowId, false)
		node.ctx.SetVariable("expires_at", "string", node.tokenExpiresAt, "", node.FlowOpCtx().FlowId, false)
	}
}

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Info(" Executing Node . Name = ", node.Meta().Label)

	if node.accessToken != "" {
		if time.Now().After(node.tokenExpiresAt) {
			node.refreshToken()
		}
	}

	var templateBuffer bytes.Buffer
	var urlTemplateBuffer bytes.Buffer
	templateParams := RestActionNodeTemplateParams{}
	templateParams.Variable = msg.Payload.Value
	templateParams.Message = msg
	templateParams.Token = node.accessToken

	node.reqTemplate.Execute(&templateBuffer, templateParams)
	node.urlTemplate.Execute(&urlTemplateBuffer, templateParams)

	node.GetLog().Debug("Url:", urlTemplateBuffer.String())
	node.GetLog().Debug("Request:", templateBuffer.String())
	req, err := http.NewRequest(node.config.Method, urlTemplateBuffer.String(), &templateBuffer)
	for i := range node.config.Headers {
		req.Header.Add(node.config.Headers[i].Name, node.config.Headers[i].Value)
	}

	if node.config.HeadersVariableName != "" {
		node.GetLog().Debug("HeadersVariableName is set")
		headers , err := node.ctx.GetVariable(node.config.HeadersVariableName,node.FlowOpCtx().FlowId)
		if err == nil {
			headersMap , ok := headers.Value.(map[string]interface{})
			if ok {
				for k := range headersMap {
					v,_ := headersMap[k].(string)
					req.Header.Add(k,v)
				}
			}else {
				node.GetLog().Debug("Can't cast header variable")
			}
		}else {
			node.GetLog().Debug("Can't load headers variable, err:",err)
		}

	}

	if err != nil {
		return []model.NodeID{node.Meta().ErrorTransition}, err
	}
	if node.config.Auth.Enabled {
		req.Header.Add("Authorization", "Bearer "+node.accessToken)
	}

	node.httpClient.CheckRedirect = func(redirRequest *http.Request, via []*http.Request) error {
		// Go's http.DefaultClient does not forward headers when a redirect 3xx
		// response is received. Thus, the header (which in this case contains the
		// Authorization token) needs to be passed forward to the redirect
		// destinations.
		redirRequest.Header = req.Header

		// Go's http.DefaultClient allows 10 redirects before returning an
		// an error. We have mimicked this default behavior.
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}

	node.GetLog().Debugf("Request headers :%v",req.Header)

	resp, err := node.httpClient.Do(req)
	if err != nil {
		return []model.NodeID{node.Meta().ErrorTransition}, err
	}
	bResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []model.NodeID{node.Meta().ErrorTransition}, err
	}
	if node.config.LogResponse {
		node.GetLog().Info(" Response:", string(bResponse))
	}

	if node.config.Auth.Enabled {
		if resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 400 {
			node.GetLog().Infof(" Done . Name = %s,Status = %s", node.Meta().Label, resp.Status)
			// Maybe token is not valid anymore , refreshing token and reDoing request.
			node.refreshToken()
			req.Header.Set("Authorization", "Bearer "+node.accessToken)
			resp, err = node.httpClient.Do(req)
			if err != nil {
				return []model.NodeID{node.Meta().ErrorTransition}, err
			}
		}
	}else {
		if resp.StatusCode >= 400 {
			return []model.NodeID{node.Meta().ErrorTransition}, err
		}
	}

	var jData interface{}
    var xTree tree.Node
	for i := range node.config.ResponseMapping {
		if node.config.ResponseMapping[i].PathType == "xml" {
			if xTree == nil {
				rReader := bytes.NewReader(bResponse)
				xTree, err = xmltree.ParseXML(rReader)
			}

			if xTree != nil {
				var varValue interface{}
				var xpExec = goxpath.MustParse(node.config.ResponseMapping[i].Path)

				switch node.config.ResponseMapping[i].TargetVariableType {
				case "string":
					result, err := xpExec.Exec(xTree)
					if err == nil {
						node.GetLog().Info("Xpath result :", result.String())
						varValue = result.String()
					}
				case "bool":
					result, err := xpExec.ExecBool(xTree)
					if err == nil {
						node.GetLog().Info("Xpath result :", result)
						varValue = result
					}
				case "int":
					result, err := xpExec.ExecNum(xTree)
					if err == nil {
						node.GetLog().Info("<Node> Xpath result :", int(result))
						varValue = int(result)
					}
				case "float":
					result, err := xpExec.ExecNum(xTree)
					if err == nil {
						node.GetLog().Info("Xpath result :", result)
						varValue = result
					}
				}

				if err != nil {
					node.GetLog().Error("Can't find result :", err)
				} else {
					flowId := node.FlowOpCtx().FlowId
					if node.config.ResponseMapping[i].IsVariableGlobal {
						flowId = "global"
					}
					node.ctx.SetVariable(node.config.ResponseMapping[i].TargetVariableName, node.config.ResponseMapping[i].TargetVariableType, varValue, "", flowId, false)
				}

			} else {
				node.GetLog().Error("Can't parse XML :", err)
			}
			//fmt.Println(res)
		} else if node.config.ResponseMapping[i].PathType == "json" {
			if jData == nil {
				err = json.Unmarshal(bResponse, &jData)
				if err != nil {
					node.GetLog().Error("JSON parsing error :",err)
					continue
				}
			}
			if jData != nil {
				varValue, err := jsonpath.JsonPathLookup(jData, node.config.ResponseMapping[i].Path)
				if err == nil {
					flowId := node.FlowOpCtx().FlowId
					if node.config.ResponseMapping[i].IsVariableGlobal {
						flowId = "global"
					}
					node.ctx.SetVariable(node.config.ResponseMapping[i].TargetVariableName, node.config.ResponseMapping[i].TargetVariableType, varValue, "", flowId, false)

				}
			}else {
				node.GetLog().Error("Emtpy json document")
			}


		}

	}



	node.GetLog().Infof(" Done . Name = %s,Status = %s", node.Meta().Label, resp.Status)
	return []model.NodeID{node.Meta().SuccessTransition}, nil
}
