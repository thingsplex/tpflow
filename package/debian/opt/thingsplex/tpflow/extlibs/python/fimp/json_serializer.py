import json
from .message import Message


class JsonSerializer:
    @staticmethod
    def fimp_msg_to_string(fimp_msg:Message)->str:
        jmsg = dict()
        jmsg["serv"] = fimp_msg.service
        jmsg["type"] = fimp_msg.msg_type
        jmsg["val_t"] = str(fimp_msg.value_type)
        jmsg["val"] = fimp_msg.value
        jmsg["tags"] = fimp_msg.tags
        jmsg["props"] = fimp_msg.props
        jmsg["ctime"] = fimp_msg.mctime
        jmsg["ver"] = fimp_msg.ver
        jmsg["uid"] = fimp_msg.uid
        return json.dumps(jmsg)

    @staticmethod
    def string_to_fimp_msg(fimp_str:str)->Message:
        msg = Message()
        jobj = json.loads(fimp_str)
        msg.service = jobj["serv"]
        msg.msg_type = jobj["type"]
        msg.value_type = jobj["val_t"]
        msg.value = jobj["val"]
        msg.tags = jobj["tags"]
        msg.props = jobj["props"]
        if "ctime" in jobj :
            msg.mctime = jobj["ctime"]
        if "uid" in jobj :
            msg.uid = jobj["uid"]
        if "corid" in jobj :
            msg.corid = jobj["corid"]
        return msg

