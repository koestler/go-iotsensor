package device

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"time"
)

type Config interface {
	Name() string
	SkipFields() []string
	SkipCategories() []string
	TelemetryViaMqttClients() []string
	RealtimeViaMqttClients() []string
	LogDebug() bool
	LogComDebug() bool
}

type Device interface {
	Name() string
	Config() Config
	Registers() dataflow.Registers
	GetRegister(registerName string) dataflow.Register
	LastUpdated() time.Time
	Model() string
	Run() error
	Shutdown()
	ShutdownChan() chan struct{}
}
