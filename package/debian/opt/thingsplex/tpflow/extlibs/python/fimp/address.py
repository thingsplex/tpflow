from enum import Enum


class PayloadType(Enum):
    DEFAULT = "j1"


class MsgType(Enum):
    CMD = "cmd"
    EVT = "evt"


class ResourceType(Enum):
    DEVICE = "dev"
    APP = "app"
    ADAPTER = "ad"


class Address:
    def __init__(self,payload_type:PayloadType=PayloadType.DEFAULT.value, msg_type:MsgType = MsgType.CMD.value, resource_type:ResourceType = ResourceType.DEVICE.value,
                 resource_name:str = "",resource_address:str="",service_name:str="",service_address:str=""):
        self.payload_type = payload_type
        self.msg_type = msg_type
        self.resource_type = resource_type
        self.resource_name = resource_name
        self.resource_address = resource_address
        self.service_name = service_name
        self.service_address = service_address

    def serialize(self)->str:
        if self.resource_type == ResourceType.DEVICE.value:
            return "pt:%s/mt:%s/rt:dev/rn:%s/ad:%s/sv:%s/ad:%s"%(self.payload_type,self.msg_type,self.resource_name,self.resource_address,self.service_name,self.service_address)

        elif self.resource_type == ResourceType.APP.value:
            return "pt:%s/mt:%s/rt:app/rn:%s/ad:%s"%(self.payload_type,self.msg_type,self.resource_name,self.resource_address)

        elif self.resource_type == ResourceType.ADAPTER.value:
            return "pt:%s/mt:%s/rt:ad/rn:%s/ad:%s"%(self.payload_type,self.msg_type,self.resource_name,self.resource_address)

    @staticmethod
    def parse(addr_str:str):
        addr = Address()
        addr_tokens = addr_str.split("/")
        for a_token in addr_tokens :
            splited_token = a_token.split(":")
            if splited_token[0] == "pt":
                addr.payload_type = splited_token[1]
            elif splited_token[0] == "mt":
                addr.msg_type = splited_token[1]
            elif splited_token[0] == "rt":
                addr.resource_type = splited_token[1]
            elif splited_token[0] == "rn":
                addr.resource_name = splited_token[1]
            elif splited_token[0] == "ad":
                addr.resource_address = splited_token[1]
            elif splited_token[0] == "sv":
                addr.service_name = splited_token[1]
            elif splited_token[0] == "ad" and addr.service_name != "":
                addr.service_address = splited_token[1]

        return addr
