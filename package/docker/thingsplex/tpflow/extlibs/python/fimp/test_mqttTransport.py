from unittest import TestCase
from .message import Message, ValueType
from .address import Address, MsgType, ResourceType
from .mqtt_transport import MqttTransport
import time


def msg_handler(topic, address, fimp_msg):
    print("Message from address :" + topic)


class MsgResponder:
    def __init__(self):
        self.mq_transport = MqttTransport()
        self.mq_transport.set_message_handler(self.msg_handler)
        self.mq_transport.connect()
        self.mq_transport.subscribe("pt:j1/mt:cmd/rt:dev/rn:pylib/ad:1/sv:out_bin_switch/ad:111")

    def msg_handler(self, topic, address, fimp_msg):
        print("---Request----")
        msg = Message("out_bin_switch", "evt.binary.report", True, ValueType.BOOL.value)
        resp_topic = "pt:j1/mt:evt/rt:dev/rn:pylib/ad:1/sv:out_bin_switch/ad:111"
        self.mq_transport.publish(resp_topic, msg)
        print("---Response was sent---")
    def stop(self):
        self.mq_transport.stop()


class TestMqttTransport(TestCase):
    def test_publish(self):
        msg = Message("out_bin_switch", "cmd.binary.set", True, ValueType.BOOL.value)
        addr = Address(msg_type=MsgType.CMD.value, resource_type=ResourceType.DEVICE.value, resource_name="pylib", resource_address="1",
                       service_name="out_bin_switch", service_address="111")
        mq_transport = MqttTransport()
        mq_transport.connect()
        # time.sleep(2)
        print("Publishing to topic = " + addr.serialize())
        mq_transport.publish(addr.serialize(), msg)
        # time.sleep(2)
        mq_transport.stop()

    def test_on_message(self):
        mq_transport = MqttTransport()
        mq_transport.set_message_handler(msg_handler)
        mq_transport.connect()
        mq_transport.subscribe("pt:j1/mt:evt/rt:dev/rn:pylib/ad:1/sv:out_bin_switch/ad:111")
        time.sleep(1)
        msg = Message("out_bin_switch", "evt.binary.report", True, ValueType.BOOL.value)
        addr = Address(msg_type=MsgType.EVT.value, resource_type=ResourceType.DEVICE.value, resource_name="pylib", resource_address="1",
                       service_name="out_bin_switch", service_address="111")
        mq_transport.publish(addr.serialize(), msg)
        time.sleep(1)
        mq_transport.stop()

    def test_send_request(self):
        responder = MsgResponder()
        mq_transport = MqttTransport()
        mq_transport.connect()
        msg = Message("out_bin_switch", "cmd.binary.set", True, ValueType.BOOL.value)
        req_topic = "pt:j1/mt:cmd/rt:dev/rn:pylib/ad:1/sv:out_bin_switch/ad:111"
        resp_topic = "pt:j1/mt:evt/rt:dev/rn:pylib/ad:1/sv:out_bin_switch/ad:111"
        response = mq_transport.send_request(req_topic, msg, resp_topic)
        print(response)
        mq_transport.stop()
        responder.stop()
