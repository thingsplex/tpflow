package node

import (
	"bytes"
	"encoding/json"
	"github.com/ChrisTrenkamp/goxpath"
	"github.com/ChrisTrenkamp/goxpath/tree/xmltree"
	"github.com/alivinco/tpflow/model"
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
type RestActionNode struct {
	BaseNode
	ctx *model.Context
	config RestActionNodeConfig
	reqTemplate *template.Template
	urlTemplate *template.Template
	httpClient  *http.Client
	accessToken string
	tokenExpiresAt time.Time
}

type ResponseToVariableMap struct {
	Name string
	Path string
	PathType string // xml , json
	TargetVariableName string
	IsVariableGlobal bool
	TargetVariableType string
	UpdateTriggerMessage bool
}

type Header struct {
	Name string
	Value string
}

type RestActionNodeConfig struct {
	Url string
	Method string // GET,POST,PUT,DELETE, etc.
	TemplateVariableName string
	IsVariableGlobal bool
	RequestPayloadType string // json,xml,string
	RequestTemplate string
	Headers []Header
	ResponseMapping []ResponseToVariableMap
	LogResponse bool
	Auth OAuth
}

type OAuth struct {
	Enabled bool
	GrantType string
	Url string
	ClientID string
	ClientSecret string
	Scope string
	Username string
	Password string

}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn int64 `json:"expires_in"`
}

type RestActionNodeTemplateParams struct {
	Variable interface{}
	Token string
	Message *model.Message
}

func NewRestActionNode(flowOpCtx *model.FlowOperationalContext,meta model.MetaNode,ctx *model.Context) model.Node {
	node := RestActionNode{ctx:ctx}
	node.meta = meta
	node.flowOpCtx = flowOpCtx
	node.config = RestActionNodeConfig{}
	node.httpClient = &http.Client{}
	node.SetupBaseNode()
	return &node
}

func (node *RestActionNode) LoadNodeConfig() error {
	err := mapstructure.Decode(node.meta.Config,&node.config)

	funcMap := template.FuncMap{
		"variable": func(varName string,isGlobal bool)(interface{},error) {
			//node.getLog().Debug("Getting variable by name ",varName)
			var vari model.Variable
			var err error
			if isGlobal {
				vari , err = node.ctx.GetVariable(varName,"global")
			}else {
				vari , err = node.ctx.GetVariable(varName,node.flowOpCtx.FlowId)
			}

			if vari.IsNumber() {
				return vari.ToNumber()
			}
			vstr , ok := vari.Value.(string)
			if ok {
				return vstr,err
			}else {
				return "",errors.New("Only simple types are supported ")
			}
		},
	}

	if err != nil{
		node.getLog().Error(" Failed while loading configurations.Error:",err)

	}else {
		node.reqTemplate,err = template.New("request").Funcs(funcMap).Parse(node.config.RequestTemplate)
		if err != nil {
			node.getLog().Error(" Failed while parsing request template.Error:",err)
		}
		node.urlTemplate,err = template.New("url").Funcs(funcMap).Parse(node.config.Url)
		if err != nil {
			node.getLog().Error(" Failed while parsing url template.Error:",err)
		}
		node.Authenticate("","")
	}
	return err
}

func (node *RestActionNode) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *RestActionNode) loadAccessTokenFromContext()bool {
	accessTokenVar,err := node.ctx.GetVariable("access_token",node.flowOpCtx.FlowId)

	if err != nil {
		node.getLog().Info(" Error , can't load access token from context.Err",err)
		node.accessToken = ""
		return false
	}
	var ok bool
	node.accessToken ,ok = accessTokenVar.Value.(string)
	if !ok {
		node.getLog().Info(" Error , can't load access token from context,variable is not a string .Err",err)
		node.accessToken = ""
		return false
	}

	if node.accessToken == "" {
		return false
	}
	return true
}


func (node *RestActionNode) Authenticate(username string,password string ) {
	if ! node.config.Auth.Enabled {
		node.accessToken = ""
		return
	}
	if node.loadAccessTokenFromContext() {
		node.getLog().Info(" Auth token already loaded ")
		return
	}
	node.getLog().Info(" Requesting new token")
	data := url.Values{}
	data.Add("grant_type", node.config.Auth.GrantType)
	data.Add("client_id", node.config.Auth.ClientID)
	data.Add("client_secret", node.config.Auth.ClientSecret)
	data.Add("scope", node.config.Auth.Scope)
	if node.config.Auth.GrantType == "authorization_code"{
		data.Add("code", node.config.Auth.Password)
	}
	if node.config.Auth.GrantType == "password"{
		data.Add("username", node.config.Auth.Username)
		data.Add("password", node.config.Auth.Password)
	}

	client := &http.Client{}
	r, _ := http.NewRequest("POST", node.config.Auth.Url, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type","application/x-www-form-urlencoded")

	//bReq ,err := ioutil.ReadAll(r.Body)
	//node.getLog().Info("<RestActionNode> Auth request:",string(bReq))

	resp, err := client.Do(r)
	if err != nil {
		node.getLog().Info(" Auth error . Error : ", err)
		return
	}
	bData ,err := ioutil.ReadAll(resp.Body)
	if err == nil {

		node.getLog().Info("Auth Response:",string(bData))

		jresp := AuthResponse{}
		json.Unmarshal([]byte(bData), &jresp)

		node.getLog().Info("Auth Response:",string(bData))

		node.accessToken = jresp.AccessToken
		node.tokenExpiresAt = time.Now().Add(time.Second* time.Duration(jresp.ExpiresIn))

		node.getLog().Infof(" Auth successfully . Access toke expires after %d seconds , at ", jresp.ExpiresIn,node.tokenExpiresAt.Format("2006-01-02 15:04:05"))
		node.ctx.SetVariable("access_token","object",jresp.AccessToken,"",node.flowOpCtx.FlowId,false )
		node.ctx.SetVariable("refresh_token","string",jresp.RefreshToken,"",node.flowOpCtx.FlowId,false )
		node.ctx.SetVariable("expires_at","string",node.tokenExpiresAt,"",node.flowOpCtx.FlowId,false )
	}

}

func (node *RestActionNode) refreshToken() {
	node.getLog().Info(" Refreshing access token . ")
	refreshTokenVar,err := node.ctx.GetVariable("refresh_token",node.flowOpCtx.FlowId)

	if err != nil {
		node.getLog().Info(" Error , can't load refresh token from context.Err",err)
		node.accessToken = ""
		return
	}
	if refreshTokenVar.Value == "" {
		node.getLog().Debug(" Refresh token is empty ")
		return
	}
	refreshToken ,ok := refreshTokenVar.Value.(string)
	if !ok {
		node.getLog().Info(" Error , can't load refresh token from context,variable is not a string .Err",err)
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
	r.Header.Add("Content-Type","application/x-www-form-urlencoded")
	resp, err := client.Do(r)
	if err != nil {
		node.getLog().Info(" Auth error . Error : ", err)
		return
	}
	bData ,err := ioutil.ReadAll(resp.Body)
	if err == nil {
		jresp := AuthResponse{}
		json.Unmarshal([]byte(bData), &jresp)

		node.getLog().Info("Auth Response:",string(bData))

		node.accessToken = jresp.AccessToken
		node.tokenExpiresAt = time.Now().Add(time.Second* time.Duration(jresp.ExpiresIn))

		node.getLog().Infof(" Token refreshed successfully . Access toke expires after %d seconds , at ", jresp.ExpiresIn,node.tokenExpiresAt.Format("2006-01-02 15:04:05"))
		node.ctx.SetVariable("access_token","string",jresp.AccessToken,"",node.flowOpCtx.FlowId,false )
		node.ctx.SetVariable("refresh_token","string",jresp.RefreshToken,"",node.flowOpCtx.FlowId,false )
		node.ctx.SetVariable("expires_at","string",node.tokenExpiresAt,"",node.flowOpCtx.FlowId,false )
	}
}

func (node *RestActionNode) OnInput( msg *model.Message) ([]model.NodeID,error) {
	node.getLog().Info(" Executing RestActionNode . Name = ", node.meta.Label)

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

	node.reqTemplate.Execute(&templateBuffer,templateParams)
	node.urlTemplate.Execute(&urlTemplateBuffer,templateParams)

	node.getLog().Debug("Url:",urlTemplateBuffer.String())
	node.getLog().Debug("Request:",templateBuffer.String())
	req, err := http.NewRequest(node.config.Method, urlTemplateBuffer.String(), &templateBuffer)
	for i := range node.config.Headers{
		req.Header.Add(node.config.Headers[i].Name,node.config.Headers[i].Value)
	}

	if err != nil {
		return []model.NodeID{node.meta.ErrorTransition},err
	}
	if node.config.Auth.Enabled {
		req.Header.Add("Authorization","Bearer "+node.accessToken)
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

	resp, err := node.httpClient.Do(req)
	if err != nil {
		return []model.NodeID{node.meta.ErrorTransition},err
	}
	if node.config.Auth.Enabled {
		if resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 400 {
			node.getLog().Infof(" Done . Name = %s,Status = %s", node.meta.Label,resp.Status)
			// Maybe token is not valid anymore , refreshing token and reDoing request.
			node.refreshToken()
			req.Header.Set("Authorization","Bearer "+node.accessToken)
			resp, err = node.httpClient.Do(req)
			if err != nil {
				return []model.NodeID{node.meta.ErrorTransition},err
			}
		}
	}


	for i := range node.config.ResponseMapping {
		if node.config.ResponseMapping[i].PathType == "xml" {
			xTree, err := xmltree.ParseXML(resp.Body)
			if err == nil {
				var varValue interface{}
				var xpExec = goxpath.MustParse(node.config.ResponseMapping[i].Path)

				switch node.config.ResponseMapping[i].TargetVariableType {
				case "string":
					result, err := xpExec.Exec(xTree)
					if err == nil {
						node.getLog().Info("Xpath result :",result.String())
						varValue = result.String()
					}
				case "bool":
					result, err := xpExec.ExecBool(xTree)
					if err == nil {
						node.getLog().Info("Xpath result :",result)
						varValue = result
					}
				case "int":
					result, err := xpExec.ExecNum(xTree)
					if err == nil {
						node.getLog().Info("<RestActionNode> Xpath result :",int(result))
						varValue = int(result)
					}
				case "float":
					result, err := xpExec.ExecNum(xTree)
					if err == nil {
						node.getLog().Info("Xpath result :",result)
						varValue = result
					}
				}

				if err != nil {
					node.getLog().Error("Can't find result :",err)
				}else {
					flowId := node.flowOpCtx.FlowId
					if node.config.ResponseMapping[i].IsVariableGlobal {
						flowId = "global"
					}
					node.ctx.SetVariable(node.config.ResponseMapping[i].TargetVariableName,node.config.ResponseMapping[i].TargetVariableType,varValue,"",flowId,false )
				}


			}else {
				node.getLog().Error("Can't parse XML :",err)
			}
			//fmt.Println(res)
		}else if node.config.ResponseMapping[i].PathType == "json" {
			var jData interface{}
			bData ,err := ioutil.ReadAll(resp.Body)
			if err == nil {
				json.Unmarshal([]byte(bData), &jData)
				varValue, err := jsonpath.JsonPathLookup(jData, node.config.ResponseMapping[i].Path)
				if err == nil {
					flowId := node.flowOpCtx.FlowId
					if node.config.ResponseMapping[i].IsVariableGlobal {
						flowId = "global"
					}
					node.ctx.SetVariable(node.config.ResponseMapping[i].TargetVariableName,node.config.ResponseMapping[i].TargetVariableType,varValue,"",flowId,false )

				}
			}
		}

	}

	if node.config.LogResponse {
		var respBuff bytes.Buffer
		respBuff.ReadFrom(resp.Body)
		node.getLog().Info(" Response:",respBuff.String())
	}

	node.getLog().Infof(" Done . Name = %s,Status = %s", node.meta.Label,resp.Status)
	return []model.NodeID{node.meta.SuccessTransition},nil
}

