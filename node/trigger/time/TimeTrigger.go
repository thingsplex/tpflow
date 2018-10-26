package time

import (
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/node/base"
	"github.com/kelvins/sunrisesunset"
	"github.com/mitchellh/mapstructure"
	"github.com/robfig/cron"
	"time"
)

const SUNRISE = "sunrise"
const SUNSET = "sunset"
const TIME_FORMAT = "2006-01-02 15:04:05"

// Time trigger node
type Node struct {
	base.BaseNode
	ctx            *model.Context
	config         NodeConfig
	cron           *cron.Cron
	astroTimer     *time.Timer
	nextAstroEvent string
	cronMessageCh  model.MsgPipeline
	msgInStream    model.MsgPipeline
}

type NodeConfig struct {
	DefaultMsg              model.Variable
	Expressions             []TimeExpression
	GenerateAstroTimeEvents bool
	Latitude                float64
	Longitude               float64
	TimeZone                float64
	SunriseTimeOffset       float64 // astro time offset in minutes
	SunsetTimeOffset        float64 // astro time offset in minutes
}

type TimeExpression struct {
	Name       string
	Expression string //https://godoc.org/github.com/robfig/cron#Job
	Comment    string
}

func NewNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := Node{ctx: ctx}
	node.SetStartNode(true)
	node.SetMsgReactorNode(true)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetMeta(meta)
	node.config = NodeConfig{}
	node.cron = cron.New()
	node.cronMessageCh = make(model.MsgPipeline)
	node.SetupBaseNode()
	return &node
}

func (node *Node) LoadNodeConfig() error {
	err := mapstructure.Decode(node.Meta().Config, &node.config)
	if err != nil {
		node.GetLog().Error("Can't load config.Err", err)
	}
	node.config.TimeZone = 1
	return err
}

// is invoked when node is started
func (node *Node) Init() error {
	if node.config.GenerateAstroTimeEvents {

	} else {
		for i := range node.config.Expressions {
			node.cron.AddFunc(node.config.Expressions[i].Expression, func() {
				node.GetLog().Debug("New time event")
				msg := model.Message{Payload: fimpgo.FimpMessage{Value: node.nextAstroEvent, ValueType: fimpgo.VTypeString},
					Header: map[string]string{"name": node.config.Expressions[i].Name}}
				node.cronMessageCh <- msg
			})

		}
		node.cron.Start()

	}

	return nil
}

func (node *Node) getSunriseAndSunset(nextDay bool) (sunrise time.Time, sunset time.Time, err error) {
	currentDate := time.Now()
	if nextDay {
		oneDayDuration := time.Duration(time.Hour * 24)
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
	sunrise = time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), sunrise.Hour(), sunrise.Minute(), 0, 0, time.Local)
	sunset = time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), sunset.Hour(), sunset.Minute(), 0, 0, time.Local)
	return
}

func (node *Node) getTimeUntilNextEvent() (eventTime time.Duration, eventType string, err error) {
	sunrise, sunset, err := node.getSunriseAndSunset(false)

	timeTillSunrise := time.Until(sunrise)
	timeTillSunset := time.Until(sunset)
	if timeTillSunrise.Minutes() > 0 && timeTillSunset.Minutes() > 0 {
		return timeTillSunrise, SUNRISE, err
	} else if timeTillSunrise.Minutes() <= 0 && timeTillSunset.Minutes() > 0 {
		return timeTillSunset, SUNSET, err
	} else {
		// sunrise and sunset are in next day .
		sunrise, sunset, err = node.getSunriseAndSunset(true)
		return time.Until(sunrise), SUNRISE, err
	}

}

func (node *Node) scheduleNextAstroEvent() {
	node.GetLog().Debug("Time now ", time.Now().Format(TIME_FORMAT))
	node.GetLog().Infof(" Scheduling next astro event at location Lat = %f,Long = %f ", node.config.Latitude, node.config.Longitude)
	sunriseTime, sunsetTime, err := node.getSunriseAndSunset(false)
	node.GetLog().Debug(" Today Sunrise is at ", sunriseTime.Format(TIME_FORMAT))
	node.GetLog().Debug(" Today Sunset is  at ", sunsetTime.Format(TIME_FORMAT))

	timeUntilEvent, eventType, err := node.getTimeUntilNextEvent()
	if err != nil {
		node.GetLog().Debugf(" Event can't be scheduled .Error:", err)
		return
	}
	node.nextAstroEvent = eventType
	node.GetLog().Debugf(" %f hours  left until next %s . Event will fire at time %s ", timeUntilEvent.Hours(), eventType, time.Now().Add(timeUntilEvent).Format(TIME_FORMAT))
	var offset float64
	if eventType == SUNRISE {
		offset = node.config.SunriseTimeOffset
	} else if eventType == SUNSET {
		offset = node.config.SunsetTimeOffset
	}
	if offset != 0 {
		timeUntilEvent = timeUntilEvent + time.Duration(offset*float64(time.Minute))
		node.GetLog().Debugf(" Applying offset . New time is %s ", time.Now().Add(timeUntilEvent).Format(TIME_FORMAT))
	}

	node.astroTimer = time.AfterFunc(timeUntilEvent, func() {
		node.GetLog().Debug(" Astro time event.Event type = ", node.nextAstroEvent)
		msg := model.Message{Payload: fimpgo.FimpMessage{Value: node.nextAstroEvent, ValueType: fimpgo.VTypeString},
			Header: map[string]string{"astroEvent": node.nextAstroEvent}}
		node.cronMessageCh <- msg

	})
	node.GetLog().Info(" Event scheduled . Type = ", node.nextAstroEvent)
}

func (node *Node) ConfigureInStream(activeSubscriptions *[]string, msgInStream model.MsgPipeline) {
	node.msgInStream = msgInStream
}

// is invoked when node flow is stopped
func (node *Node) Cleanup() error {
	if node.config.GenerateAstroTimeEvents {
		node.astroTimer.Stop()
	} else {
		node.cron.Stop()
	}
	return nil
}

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	return nil, nil
}

func (node *Node) WaitForEvent(nodeEventStream chan model.ReactorEvent) {
	node.SetReactorRunning(true)
	defer func() {
		node.SetReactorRunning(false)
		node.GetLog().Debug(" Reactor-WaitForEvent is stopped ")
	}()
	for {
		if node.config.GenerateAstroTimeEvents {
			node.scheduleNextAstroEvent()
		}

		select {
		case newMsg := <-node.cronMessageCh:
			newEvent := model.ReactorEvent{Msg: newMsg, TransitionNodeId: node.Meta().SuccessTransition}
			node.FlowRunner()(newEvent)
		case signal := <-node.FlowOpCtx().NodeControlSignalChannel:
			node.GetLog().Debug("Control signal ")
			if signal == model.SIGNAL_STOP {
				return
			}
		}

	}

}
