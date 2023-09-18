package mqttClient

import (
	"net/url"
	"time"
)

type Config interface {
	Name() string
	Broker() *url.URL

	User() string
	Password() string
	ClientId() string

	Qos() byte
	KeepAlive() time.Duration
	ConnectRetryDelay() time.Duration
	ConnectTimeout() time.Duration
	TopicPrefix() string
	MaxBacklogSize() int

	AvailabilityEnabled() bool
	AvailabilityTopic() string
	AvailabilityRetain() bool

	TelemetryEnabled() bool
	TelemetryTopic() string
	TelemetryInterval() time.Duration
	TelemetryRetain() bool

	StructureEnabled() bool
	StructureTopic() string
	StructureInterval() time.Duration
	StructureRetain() bool

	RealtimeEnabled() bool
	RealtimeTopic() string
	RealtimeInterval() time.Duration
	RealtimeRepeat() bool
	RealtimeRetain() bool

	LogDebug() bool
	LogMessages() bool
}

type Client interface {
	Name() string
	Config() Config
	Run()
	Shutdown()
	Publish(topic string, payload []byte, qos byte, retain bool)
	AddRoute(subscribeTopic string, messageHandler MessageHandler)
}

type MessageHandler func(Message)

type Message struct {
	topic   string
	payload []byte
}

func (m Message) Topic() string {
	return m.topic
}

func (m Message) Payload() []byte {
	return m.payload
}
