package node

import (
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/model"
	"github.com/kelvins/sunrisesunset"
	"github.com/mitchellh/mapstructure"
	"github.com/robfig/cron"
	"time"
)

const  SUNRISE = "sunrise"
const  SUNSET  = "sunset"
const  TIME_FORMAT = "2006-01-02 15:04:05"

type TimeTriggerNode struct {
	BaseNode
	ctx            *model.Context
	config         TimeTriggerConfig
	cron           *cron.Cron
	astroTimer     * time.Timer
	nextAstroEvent string
	cronMessageCh  model.MsgPipeline
	msgInStream    model.MsgPipeline
}

type TimeTriggerConfig struct {
	DefaultMsg model.Variable
	Expressions []TimeExpression
	GenerateAstroTimeEvents bool
	Latitude float64
	Longitude float64
	TimeZone float64
	SunriseTimeOffset float64 // astro time offset in minutes
	SunsetTimeOffset float64 // astro time offset in minutes
}

type TimeExpression struct {
	Name string
	Expression string   //https://godoc.org/github.com/robfig/cron#Job
	Comment string
}

func NewTimeTriggerNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context, transport *fimpgo.MqttTransport) model.Node {
	node := TimeTriggerNode{ctx: ctx}
	node.isStartNode = true
	node.isMsgReactor = true
	node.flowOpCtx = flowOpCtx
	node.meta = meta
	node.config = TimeTriggerConfig{}
	node.cron = cron.New()
	node.cronMessageCh = make(model.MsgPipeline)
	node.SetupBaseNode()
	return &node
}

func (node *TimeTriggerNode) LoadNodeConfig() error {
	err := mapstructure.Decode(node.meta.Config,&node.config)
	if err != nil{
		node.getLog().Error("Can't load config.Err",err)
	}
	node.config.TimeZone = 1
	return err
}


// is invoked when node is started
func (node *TimeTriggerNode) Init() error {
	if node.config.GenerateAstroTimeEvents {


	}else {
		for i := range node.config.Expressions {
			node.cron.AddFunc(node.config.Expressions[i].Expression,func() {
				node.getLog().Debug("New time event")
				msg := model.Message{Payload:fimpgo.FimpMessage{Value:node.nextAstroEvent,ValueType:fimpgo.VTypeString},
					Header:map[string]string{"name":node.config.Expressions[i].Name}}
				node.cronMessageCh <- msg
			})

		}
		node.cron.Start()

	}

	return nil
}

func (node *TimeTriggerNode) getSunriseAndSunset(nextDay bool) (sunrise time.Time,sunset time.Time,err error) {
	currentDate := time.Now()
	if nextDay {
		oneDayDuration := time.Duration(time.Hour*24)
		currentDate = currentDate.Add(oneDayDuration)
	}

	p := sunrisesunset.Parameters{
		Latitude:  node.config.Latitude,
		Longitude: node.config.Longitude,
		UtcOffset: node.config.TimeZone,
		Date:      currentDate,
	}
	// Calculate the sunrise and sunset times
	sunrise, sunset, err = p.GetSunriseSunset()
	sunrise = time.Date(currentDate.Year(),currentDate.Month(),currentDate.Day(),sunrise.Hour(),sunrise.Minute(),0,0,time.Local)
	sunset = time.Date(currentDate.Year(),currentDate.Month(),currentDate.Day(),sunset.Hour(),sunset.Minute(),0,0,time.Local)
	return
}

func (node *TimeTriggerNode) getTimeUntilNextEvent() (eventTime time.Duration,eventType string,err error) {
	sunrise,sunset,err := node.getSunriseAndSunset(false)

	timeTillSunrise := time.Until(sunrise)
	timeTillSunset := time.Until(sunset)
	if timeTillSunrise.Minutes() > 0 && timeTillSunset.Minutes() > 0 {
		return timeTillSunrise,SUNRISE ,err
	}else if timeTillSunrise.Minutes() <= 0 && timeTillSunset.Minutes() > 0 {
		return timeTillSunset ,SUNSET,err
	}else {
		// sunrise and sunset are in next day .
		sunrise,sunset,err = node.getSunriseAndSunset(true)
		return time.Until(sunrise),SUNRISE,err
	}

}

func (node *TimeTriggerNode) scheduleNextAstroEvent() {
	node.getLog().Debug("Time now ",time.Now().Format(TIME_FORMAT))
	node.getLog().Infof(" Scheduling next astro event at location Lat = %f,Long = %f ",node.config.Latitude,node.config.Longitude)
	sunriseTime ,sunsetTime,err := node.getSunriseAndSunset(false)
	node.getLog().Debug(" Today Sunrise is at ",sunriseTime.Format(TIME_FORMAT))
	node.getLog().Debug(" Today Sunset is  at ",sunsetTime.Format(TIME_FORMAT))

	timeUntilEvent , eventType ,err := node.getTimeUntilNextEvent()
    if err != nil {
		node.getLog().Debugf(" Event can't be scheduled .Error:",err)
    	return
	}
	node.nextAstroEvent = eventType
	node.getLog().Debugf(" %f hours  left until next %s . Event will fire at time %s ",timeUntilEvent.Hours(),eventType,time.Now().Add(timeUntilEvent).Format(TIME_FORMAT))
	var offset float64
	if eventType == SUNRISE {
		offset = node.config.SunriseTimeOffset
	}else if eventType == SUNSET {
		offset = node.config.SunsetTimeOffset
	}
	if offset != 0 {
		timeUntilEvent = timeUntilEvent+time.Duration(offset*float64(time.Minute))
		node.getLog().Debugf(" Applying offset . New time is %s ",time.Now().Add(timeUntilEvent).Format(TIME_FORMAT))
	}

	node.astroTimer = time.AfterFunc(timeUntilEvent, func() {
		node.getLog().Debug(" Astro time event.Event type = ",node.nextAstroEvent)
		msg := model.Message{Payload:fimpgo.FimpMessage{Value:node.nextAstroEvent,ValueType:fimpgo.VTypeString},
			Header:map[string]string{"astroEvent":node.nextAstroEvent}}
		node.cronMessageCh <- msg

	})
	node.getLog().Info(" Event scheduled . Type = ",node.nextAstroEvent)
}

func (node *TimeTriggerNode) ConfigureInStream(activeSubscriptions *[]string,msgInStream model.MsgPipeline) {
	node.msgInStream = msgInStream
}

// is invoked when node flow is stopped
func (node *TimeTriggerNode) Cleanup() error {
	if node.config.GenerateAstroTimeEvents {
		node.astroTimer.Stop()
	}else {
		node.cron.Stop()
	}
	return nil
}

func (node *TimeTriggerNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	return nil,nil
}

func (node *TimeTriggerNode) WaitForEvent(nodeEventStream chan model.ReactorEvent) {
	node.isReactorRunning = true
	defer func() {
		node.isReactorRunning = false
		node.getLog().Debug(" Reactor-WaitForEvent is stopped ")
	}()

	if node.config.GenerateAstroTimeEvents {
		node.scheduleNextAstroEvent()
	}
	newMsg :=<- node.cronMessageCh

	newEvent := model.ReactorEvent{Msg:newMsg,TransitionNodeId:node.meta.SuccessTransition}
	select {
	case nodeEventStream <- newEvent:
		return
	case msg := <- node.msgInStream:
		if msg.CancelOp {
			node.cron.Stop()
			return
		}
	default:
		node.getLog().Debug(" Message is dropped (no listeners) ")
	}

}
