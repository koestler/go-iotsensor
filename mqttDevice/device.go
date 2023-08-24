package mqttDevice

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"log"
	"strings"
	"sync"
)

type Config interface {
	MqttTopics() []string
	MqttClients() []string
}

type DeviceStruct struct {
	device.State
	mqttConfig Config

	mqttClientPool *pool.Pool[mqttClient.Client]

	registers      map[string]dataflow.Register
	registersMutex sync.RWMutex
}

func CreateDevice(
	deviceConfig device.Config,
	mqttConfig Config,
	stateStorage *dataflow.ValueStorageInstance,
	mqttClientPool *pool.Pool[mqttClient.Client],
) *DeviceStruct {
	return &DeviceStruct{
		State: device.CreateState(
			deviceConfig,
			stateStorage,
		),
		mqttConfig:     mqttConfig,
		mqttClientPool: mqttClientPool,
		registers:      make(map[string]dataflow.Register),
	}
}

func (c *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	// setup mqtt listeners
	counter := 0
	for _, mc := range c.mqttClientPool.GetByNames(c.mqttConfig.MqttClients()) {
		for _, topic := range c.mqttConfig.MqttTopics() {
			log.Printf("mqttDevice[%s] subscribe to mqttClient=%s topic=%s", c.Name(), mc.Config().Name(), topic)
			mc.AddRoute(topic, func(m mqttClient.Message) {
				registerName, err := parseTopic(m.Topic())
				if err != nil {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse topic: %s", c.Name(), mc.Config().Name(), err)
					return
				}
				realtimeMessage, err := parsePayload(m.Payload())
				if err != nil {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse payload: %s", c.Name(), mc.Config().Name(), err)
					return
				}

				register := c.addIgnoreRegister(registerName, realtimeMessage)
				if register != nil {
					if v := realtimeMessage.NumericValue; v != nil {
						c.StateStorage().Fill(dataflow.NewNumericRegisterValue(c.Name(), register, *v))
					} else if v := realtimeMessage.TextValue; v != nil {
						c.StateStorage().Fill(dataflow.NewTextRegisterValue(c.Name(), register, *v))
					}
					c.SetLastUpdatedNow()
				}
			})
			counter += 1
		}
	}

	if counter < 1 {
		log.Printf("mqttDevice[%s]: no listener was starrted", c.Name())
	}

	<-ctx.Done()
	return nil, false
}

func parseTopic(topic string) (registerName string, err error) {
	registerName = topic[strings.LastIndex(topic, "/")+1:]
	if len(registerName) < 1 {
		err = fmt.Errorf("cannot extract registerName from topic='%s'", topic)
	}

	return
}

func parsePayload(payload []byte) (msg device.RealtimeMessage, err error) {
	err = json.Unmarshal(payload, &msg)
	return
}

func (c *DeviceStruct) Registers() []dataflow.Register {
	c.registersMutex.RLock()
	defer c.registersMutex.RUnlock()

	ret := make([]dataflow.Register, len(c.registers))
	i := 0
	for _, r := range c.registers {
		ret[i] = r
		i += 1
	}
	return ret
}

func (c *DeviceStruct) GetRegister(registerName string) dataflow.Register {
	c.registersMutex.RLock()
	defer c.registersMutex.RUnlock()

	if r, ok := c.registers[registerName]; ok {
		return r
	}
	return nil
}

func (c *DeviceStruct) addIgnoreRegister(registerName string, msg device.RealtimeMessage) dataflow.Register {
	// check if this register exists already and the properties are still the same
	c.registersMutex.RLock()
	if r, ok := c.registers[registerName]; ok {
		if r.Category() == msg.Category &&
			r.Description() == msg.Description &&
			r.Unit() == msg.Unit &&
			r.Sort() == msg.Sort {
			c.registersMutex.RUnlock()
			return r
		}
	}
	c.registersMutex.RUnlock()

	// check if register is on ignore list
	if device.IsExcluded(registerName, msg.Category, c.Config()) {
		return nil
	}

	// create new register
	var r dataflow.Register
	var registerType = dataflow.TextRegister

	if msg.NumericValue != nil {
		registerType = dataflow.NumberRegister
	}

	r = dataflow.CreateRegisterStruct(
		msg.Category,
		registerName,
		msg.Description,
		registerType,
		nil,
		msg.Unit,
		msg.Sort,
		false,
	)

	// add the register into the list
	c.registersMutex.Lock()
	defer c.registersMutex.Unlock()

	c.registers[registerName] = r
	return r
}

func (c *DeviceStruct) Model() string {
	return "mqtt"
}
