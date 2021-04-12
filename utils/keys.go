package utils

type ContextKey string

func (c ContextKey) String() string {
	return string(c)
}

const (
	// HTTP
	KeyHTTPStatusCode ContextKey = "http.status_code"
	KeyHTTPMethod     ContextKey = "http.method"
	KeyHTTPUrl        ContextKey = "http.url"
	KeyHTTPPath       ContextKey = "http.path"
	KeyHTTPBodySize   ContextKey = "http.body_size"
	KeyHTTPClientIP   ContextKey = "http.client_ip"
	KeyHTTPLatency    ContextKey = "http.latency"
	KeyError          ContextKey = "error"
	KeyErrorMessage   ContextKey = "error.message"

	//Kafka
	KeyKafkaTopic         ContextKey = "kafka.message.topic"
	KeyKafkaMessageLength ContextKey = "kafka.message.length"
	KeyKafkaPartition     ContextKey = "kafka.partition"
	KeyKafkaOffset        ContextKey = "kafka.offset"
)
