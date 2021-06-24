from enum import Enum
import uuid
import datetime

class ValueType(Enum):
    STRING = "string"
    INT = "int"
    FLOAT = "float"
    BOOL = "bool"
    NULL = "null"
    STR_ARRAY = "str_array"
    INT_ARRAY = "int_array"
    FLOAT_ARRAY = "float_array"
    BOOL_ARRAY = "bool_array"
    STR_MAP = "str_map"
    INT_MAP = "int_map"
    FLOAT_MAP = "float_map"
    BOOL_MAP = "bool_map"
    OBJECT = "object"


class Message:
    def __init__(self, service="", msg_type="", value=None, value_type: ValueType="string", props={}, request_msg=None,ctime=None,uid:str=""):
        self.service = service
        self.msg_type = msg_type
        self.value = value
        self.value_type = value_type
        self.tags = []
        self.props = props
        if ctime:
            self.mctime = ctime
        else:
            self.mctime = self.get_timestamp()

        if uid:
            self.uid = uid
        else:
            self.uid = str(uuid.uuid4())
        self.ver = "1.0"
        self.corid = ""

        if request_msg:
            self.corid = request_msg.uid

    def __str__(self):
        return "service = %s , msg_type = %s , val_type = %s \n value = %s  \n props : %s \n uuid : %s \n corid : %s \n"%(
            self.service, self.msg_type, self.value_type , self.value , self.props , self.uid ,self.corid
        )

    @staticmethod
    def get_uuid():
        return str(uuid.uuid4())

    @staticmethod
    def get_timestamp():
        return datetime.datetime.now().isoformat()



