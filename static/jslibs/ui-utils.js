

function getStateNameFromMessage(msg) {
    // pt:j1/mt:evt/rt:dev/rn:tibber/ad:1/sv:meter_elec/ad:12a889be-095a-47dd-8045-da84b57ae41d
    // "type": "evt.meter_ext.report",
    // state name : meter_ext@rt:dev/rn:tibber/ad:1/sv:meter_elec/ad:12a889be-095a-47dd-8045-da84b57ae41d
    let topic = msg.topic.replace("pt:j1/mt:evt/","")
    let typeS = msg.type.split(".")
    return typeS[1]+"@"+topic
}

function updateTraceByChartId(chartId,values) {
    let update = {
        x:  [[]],
        y: [[]],
    }
    for (let val of values) {
        update.x[0].push(val[0]*1000)
        update.y[0].push(val[1])
    }
    console.log("Updating TS trace for "+chartId);
    Plotly.extendTraces(chartId, update,[0]);
}

class TpFlowClient {
    constructor() {
        this.reconnectCounter = 0;
        this.ws = null ;
    }
    // Returns full URL for flow websocket
    getFlowWsUrl() {
        let flowWsUrl = "";
        if(location.protocol=="https:") {
            let pathSp = location.pathname.split("/");
            flowWsUrl = "wss://";
            return flowWsUrl+location.host+"/cloud/"+pathSp[2]+"/flow/"+flow_id+"/ws"+location.search
        }else {
            flowWsUrl = "ws://";
            return flowWsUrl+location.host+"/flow/"+flow_id+"/ws"
        }
    }

    // Returns full URL for API component
    getHttpApiUrl(comp) {
        if(location.protocol=="https:") {
            let pathSp = location.pathname.split("/");
            return location.origin+"/cloud/"+pathSp[2]+comp+location.search
        }else {
            return location.origin+comp
        }
    }

    // Returns full registry structure with states
    async loadFullStructAndStates() {
        const response = await fetch(this.getHttpApiUrl('/api/flow/context/full_struct_and_states'))
        if (response.status == 200) {
            let data = await response.json()
            return data
        }else {
            console.log("Struct and states API returned error code = ",response.status)
        }
    }

    // Returns timeseries
    async loadTimeseries(query) {
        const response = await fetch(this.getHttpApiUrl('/flow/timeseries/rest'), {method: 'post',body: JSON.stringify(query)})
        if (response.status == 200) {
            let data = await response.json()
            return data
        }else {
            console.log("Struct and states API returned error code = ",response.status)
        }
    }

    // Establishes WS connection and configures message handler
    configureWs(msgHandler) {
        this.ws = new WebSocket(this.getFlowWsUrl());
        this.ws.onopen = function() {};
        this.ws.onmessage = function (evt) {
            let msg = JSON.parse(evt.data);
            msgHandler(msg)
        };
        this.ws.onclose = function() {
            setTimeout(function() {configureWs();}, 1000);
            this.reconnectCounter++;
        };
    }

    wsPublish(msg) {
        this.ws.send(JSON.stringify(msg));
    }

}

class GlobalStructAndState {
    constructor(struct) {
        this.struct = struct;
    }

    updateStructAndState() {

    }
    // method should be invoked when
    updateStateByTopic(topic,state) {

    }

    getDevicByTopic(topic) {

    }

}




