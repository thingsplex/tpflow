from paho.mqtt import publish
from paho.mqtt import client
from .message import Message
from .address import Address
from .json_serializer import JsonSerializer
import threading
import time

def simple_publish(address: str, msg: Message, hostname, port):
    publish.single(address, JsonSerializer.fimp_msg_to_string(msg), hostname=hostname, port=port, qos=1)


class MqttTransport:
    def __init__(self, hostname: str = "localhost", port: int = 1883, qos: int = 1, client_id=None, keepalive=60, clean_session=True):
        self.hostname = hostname
        self.port = port
        self.qos = qos
        self.keepalive = keepalive
        self.client_id = client_id
        self.clean_session = clean_session
        self.mqc = client.Client(client_id=client_id, clean_session=clean_session)
        self.sub_list = []
        self.msg_handler = None
        self.active_requests = {}
        self.is_resp_listener_running = False

    def connect(self):
        self.mqc.on_connect = self.on_connect
        self.mqc.on_message = self.on_message
        self.mqc.connect(self.hostname, self.port, keepalive=self.keepalive)
        self.mqc.loop_start()

    def stop(self):
        self.mqc.loop_stop()

    def on_connect(self, mqttc, obj, flags, rc):
        print("Connected rc: " + str(rc))
        for sub in self.sub_list:
            self.mqc.subscribe(sub["topic"], sub["qos"])

    # handler format self.msg_handler(msg.topic, address, fimp_msg)
    def set_message_handler(self, handler_func):
        self.msg_handler = handler_func

    def subscribe(self, topic, qos=-1):
        if qos == -1:
            qos = self.qos
        self.sub_list.append({"topic": topic, "qos": qos})
        self.mqc.subscribe(topic, qos)

    def publish(self, topic, msg: Message, qos=-1):
        msg_str = JsonSerializer.fimp_msg_to_string(msg)
        if qos == -1:
            qos = self.qos
        self.mqc.publish(topic, msg_str, qos=qos, retain=False)

    def on_message(self, client, userdata, msg):
        print("Message from topic %s" % msg.topic)
        # print("Message from payload %s" % (str(msg.payload)))
        address = Address.parse(msg.topic)
        fimp_msg = JsonSerializer.string_to_fimp_msg(msg.payload.decode('utf-8',"ignore"))
        if self.msg_handler:
            self.msg_handler(msg.topic, address, fimp_msg)
        print("Fimp service = %s , msg type = %s"%(fimp_msg.service,fimp_msg.msg_type))
        if len(self.active_requests) > 0:
            for key, val in self.active_requests.items():
                if val["resp_topic"] == msg.topic:
                    print("Topic match")
                    if (val["resp_service"] == "" or val["resp_service"] == fimp_msg.service) and (
                            val["resp_msg_type"] == "" or val["resp_msg_type"] == fimp_msg.msg_type):
                        val["resp_msg"] = fimp_msg
                        print("Event is set")
                        val["event"].set()
                        break

    def send_request(self, req_topic, req_msg: Message, resp_topic: str, resp_service: str = "", resp_msg_type: str = "", use_corid: bool = False,
                     timeout: int = 60):
        event = threading.Event()
        self.active_requests[req_msg.uid] = {"resp_topic": resp_topic, "resp_service": resp_service, "resp_msg_type": resp_msg_type, "event": event,
                                            "resp_msg": None}
        self.subscribe(resp_topic)
        # time.sleep(1)
        self.publish(req_topic, req_msg)
        if event.wait(timeout):
            resp_msg = self.active_requests[req_msg.uid]["resp_msg"]
            del self.active_requests[req_msg.uid]
            return resp_msg
        del self.active_requests[req_msg.uid]
        return None
