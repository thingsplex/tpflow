package tpflow

type Configs struct {
	InstanceAddress       string `json:"instance_address"`
	MqttServerURI         string `json:"mqtt_server_uri"`
	MqttUsername          string `json:"mqtt_server_username"`
	MqttPassword          string `json:"mqtt_server_password"`
	MqttTopicGlobalPrefix string `json:"mqtt_topic_global_prefix"`
	FlowStorageDir        string `json:"flow_storage_dir"`
	ConnectorStorageDir   string `json:"connector_storage_dir"`
	RegistryDbFile        string `json:"registry_db_file"`
	ContextStorageDir     string `json:"context_storage_dir"`
	ExternalLibsDir       string `json:"ext_libs_dir"`
	MqttClientIdPrefix    string `json:"mqtt_client_id_prefix"`
	LogFile               string `json:"log_file"`
	LogLevel              string `json:"log_level"`
	LogFormat             string `json:"log_format"`
	IsDevMode             bool   `json:"is_dev_mode"`
	RegistryBackend       string `json:"registry_backend"`
}
