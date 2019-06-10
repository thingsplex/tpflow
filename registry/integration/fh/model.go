package fh
//
//import (
//"encoding/json"
//"fmt"
//"time"
//)
//
//type VinculumMsg struct {
//	Ver string `json:"ver"`
//	Msg Msg    `json:"msg"`
//}
//
//type VinculumMsgRequest struct {
//	Ver string     `json:"ver"`
//	Msg MsgRequest `json:"msg"`
//}
//
//type Msg struct {
//	Type    string 		`json:"type"`
//	Src     string 		`json:"src"`
//	Dst     string 		`json:"dst"`
//	Data    Data   		`json:"data"`
//	User 	UserInfo  	`json:"user,omitempty"`
//	Request DataRequest
//}
//
//type MsgRequest struct {
//	Type string      `json:"type"`
//	Src  string      `json:"src"`
//	Dst  string      `json:"dst"`
//	Data DataRequest `json:"data"`
//}
//
//type Data struct {
//	Errors     interface{}     `json:"errors"`
//	Cmd        string          `json:"cmd"`
//	Component  string          `json:"component"`
//	ParamRaw   json.RawMessage `json:"param"`
//	ChangesRaw json.RawMessage `json:"changes"`
//	Param      Param           `json:"-"`
//	RequestID  interface{}     `json:",requestId"`
//	Success    bool            `json:"success"`
//	Id         interface{} 	   `json:"id,omitempty"`
//}
//
//type DataRequest struct {
//	Errors    interface{} `json:"errors"`
//	Cmd       string      `json:"cmd"`
//	Component interface{} `json:"component"`
//	Param     Param 	  `json:"param"`
//	RequestID interface{} `json:"requestId"`
//	Success   bool        `json:"success"`
//	Id        int 		  `json:"id,omitempty"`
//}
//
//type Param struct {
//	Id         int        `json:"id,omitempty"`
//	Components []string   `json:"components,omitempty"`
//	Device     []Device   `json:"device,omitempty"`
//	Thing      []Thing    `json:"thing,omitempty"`
//	Room       []Room     `json:"room,omitempty"`
//	House      House      `json:"house,omitempty"`
//	Area       []Area     `json:"area,omitempty"`
//	Shortcut   []Shortcut `json:"shortcut,omitempty"`
//	Problem    bool       `json:"problem,omitempty"`
//	Previous   string	  `json:"prev"`					// Used for home modes and other user triggered events from the app
//	Current	   string	  `json:"current"`				// Used for home modes and other user triggered events from the app
//	Rn 		   string	  `json:"rn,omitempty"`
//	Autolock   string	  `json:"autolock,omitempty"`
//	Power	   string	  `json:"power,omitempty"`
//	Position   int	  	  `json:"position,omitempty"`	// timeline, used for shutter
//	MoveState  string	  `json:"moveState,omitempty"`  // timeline, used for shutter
//	DimValue   int	      `json:"dimValue,omitempty"`	// timeline, used for dimmers
//}
//
//type Fimp struct {
//	Adapter string `json:"adapter"`
//	Address string `json:"address"`
//	Group   string `json:"group"`
//}
//
//type Client struct {
//	Name string `json:"name"`
//}
//
//type Device struct {
//	Fimp          Fimp                   `json:"fimp"`
//	Client        Client                 `json:"client"`
//	Functionality string                 `json:"functionality"`
//	Service 	  map[string]Service	 `json:"services"`
//	ID            int                    `json:"id"`
//	Lrn           bool                   `json:"lrn"`
//	Model         string                 `json:"model"`
//	Param         map[string]interface{} `json:"param"`
//	Problem       bool                   `json:"problem"`
//	Room          int                    `json:"room"`
//	Changes       map[string]interface{}
//}
//
//type Thing struct {
//	ID      int               `json:"id"`
//	Address string            `json:"addr"`
//	Name    string            `json:"name"`
//	Devices []int             `json:"devices,omitempty"`
//	Props   map[string]string `json:"props,omitempty"`
//	RoomID  int               `json:"room"`
//}
//
//type House struct {
//	Learning interface{} `json:"learning"`
//	Mode     string      `json:"mode"`
//	Time     time.Time   `json:"time"`
//}
//
//type Room struct {
//	ID     	int         `json:"id"`
//	Param  	RoomParams 	`json:"param"`
//	Client 	Client      `json:"client"`
//	Type   	string      `json:"type"`
//	Area   	int         `json:"area,omitempty"`
//	Outside	bool        `json:"outside"`
//}
//
//type RoomParams struct {
//	Heating RoomHeating		`json:"heating"`
//}
//
//type RoomHeating struct {
//	Desired float64	`json:"desired"`
//	Target float64	`json:"target"`
//}
//
//type Service struct {
//	Addr 		string 					`json:"addr,omitempty"`
//	Enabled 	bool 					`json:"enabled,omitempty"`
//	Interfaces	[]string 				`json:"intf"`
//	Props 		map[string]interface{}	`json:"props"`
//}
//
//type UserInfo struct {
//	UID  string   `json:"uuid,omitempty"`
//	Name UserName `json:"name,omitempty"`
//}
//
//type UserName struct {
//	Fullname string `json:"fullname,omitempty"`
//}
//
//type Area struct {
//	ID   int    `json:"id"`
//	Name string `json:"name"`
//	Type string `json:"type"`
//}
//
//type Shortcut struct {
//	ID     int    `json:"id"`
//	Client Client `json:"client"`
//}
//
//// Used to avoid recursion in UnmarshalJSON below.
//type msg Msg
//
//func (m *Msg) UnmarshalJSON(b []byte) (err error) {
//	jmsg := msg{}
//	if err := json.Unmarshal(b, &jmsg); err == nil {
//		*m = Msg(jmsg)
//
//		// Parsing the request if found
//		if m.Type == "request" {
//			reqMsg := MsgRequest{}
//
//			if err := json.Unmarshal(b, &reqMsg); err == nil {
//				m.Request = MsgRequest(reqMsg).Data
//			}
//		}
//		return m.UnmarshalDataParam()
//	}
//	return nil
//}
//
//func (m *Msg) UnmarshalDataParam() error {
//	if m.Type == "notify"{
//		switch m.Data.Component {
//		case "device":
//			devices := []Device{{}}
//			if err := json.Unmarshal(m.Data.ParamRaw, &devices[0]); err != nil {
//				return err
//			}
//
//			if json.Unmarshal(m.Data.ChangesRaw, &devices[0].Changes) != nil {
//				// don't have any changes, no need to handle it anyway
//			}
//			m.Data.Param.Device = devices
//
//		case "area":
//			areas := []Area{{}}
//			if err := json.Unmarshal(m.Data.ParamRaw, &areas[0]); err != nil {
//				return err
//			}
//			m.Data.Param.Area = areas
//
//		case "house":
//			house := House{}
//			if err := json.Unmarshal(m.Data.ParamRaw, &house); err != nil {
//				return err
//			}
//			m.Data.Param.House = house
//
//		case "room":
//			rooms := []Room{{}}
//			if err := json.Unmarshal(m.Data.ParamRaw, &rooms[0]); err != nil {
//				return err
//			}
//			m.Data.Param.Room = rooms
//
//		case "shortcut":
//			shortcut := []Shortcut{{}}
//			if err := json.Unmarshal(m.Data.ParamRaw, &shortcut[0]); err != nil {
//				return err
//			}
//			m.Data.Param.Shortcut = shortcut
//
//		case "hub":
//			var param Param
//			if err := json.Unmarshal([]byte(m.Data.ParamRaw), &param); err != nil {
//				fmt.Println("Unmarshal Error ", err)
//				return err
//			}
//
//			m.Data.Param = param
//		}
//	} else {
//		var param Param
//		if err := json.Unmarshal([]byte(m.Data.ParamRaw), &param); err != nil {
//			// Param will fail for app specific request(s). We are providing paramRaw and
//			// applications has to know how to handle it. It's impossible to predict every case here
//			m.Data.Param = Param{}
//			//fmt.Println("Unmarshal Error ", err)
//		} else {
//			m.Data.Param = param
//		}
//	}
//	return nil
//}