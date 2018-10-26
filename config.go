package tpflow

type Configs struct {
	MqttServerURI         string `json:"mqtt_server_uri"`
	MqttUsername          string `json:"mqtt_server_username"`
	MqttPassword          string `json:"mqtt_server_password"`
	MqttTopicGlobalPrefix string `json:"mqtt_topic_global_prefix"`
	FlowStorageDir        string `json:"flow_storage_dir"`
	ConnectorStorageDir   string `json:"connector_storage_dir"`
	RegistryDbFile        string `json:"registry_db_file"`
	ContextStorageDir     string `json:"context_storage_dir"`
	MqttClientIdPrefix    string `json:"mqtt_client_id_prefix"`
	LogFile               string `json:"log_file"`
	LogLevel              string `json:"log_level"`
	LogFormat             string `json:"log_format"`
}
