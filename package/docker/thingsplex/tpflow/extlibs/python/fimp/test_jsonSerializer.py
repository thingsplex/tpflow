from unittest import TestCase
from .message import Message,ValueType
from .json_serializer import JsonSerializer


class TestJsonSerializer(TestCase):
    def test_fimp_msg_to_string(self):
        msg = Message(service="out_bin_switch",msg_type="cmd.binary.set",value=True,value_type=ValueType.BOOL.value)
        str_msg = JsonSerializer.fimp_msg_to_string(msg)
        print(str_msg)

    def test_string_to_fimp_msg(self):
        str_msg = '{"serv": "out_bin_switch", "type": "cmd.binary.set", "val_t": "bool", "val": true, "tags": [], "props": {}, "ctime": "2018-04-04T10:24:50.439003", "ver": "1.0", "uid": "21c31425-56fd-40ff-97d1-c6b05b79762f"}'
        msg = JsonSerializer.string_to_fimp_msg(str_msg)
        print(msg)