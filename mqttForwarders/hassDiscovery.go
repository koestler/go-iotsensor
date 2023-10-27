package mqttForwarders

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"log"
	"strings"
	"time"
)

type homeassistantDiscoveryAvailabilityStruct struct {
	Topic string `json:"t"`
}

type homeassistantDiscoveryMessage struct {
	UniqueId string `json:"uniq_id"`
	Name     string `json:"name"`

	StateTopic        string                                     `json:"stat_t"`
	Availability      []homeassistantDiscoveryAvailabilityStruct `json:"avty"`
	AvailabilityMode  string                                     `json:"avty_mode"`
	ValueTemplate     string                                     `json:"val_tpl"`
	UnitOfMeasurement string                                     `json:"unit_of_meas,omitempty"`
}

func runHomeassistantDiscoveryForwarder(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	registerFilter dataflow.RegisterFilterConf,
) {
	filter := createRegisterValueFilter(registerFilter)

	if cfg.HomeassistantDiscovery().Interval() <= 0 {
		go homeassistantDiscoveryOnUpdateModeRoutine(ctx, cfg, dev, mc, filter)
	} else {
		go homeassistantDiscoveryPeriodicModeRoutine(ctx, cfg, dev, mc, filter)
	}
}

func homeassistantDiscoveryOnUpdateModeRoutine(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	filter dataflow.RegisterFilterFunc,
) {
	if cfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->homeassistantDiscovery: start on-update mode",
			dev.Name(), mc.Name(),
		)

		defer log.Printf(
			"device[%s]->mqttClient[%s]->homeassistantDiscovery: exit",
			dev.Name(), mc.Name(),
		)
	}

	regSubscription := dev.RegisterDb().Subscribe(ctx, filter)

	for {
		select {
		case <-ctx.Done():
			return
		case reg := <-regSubscription:
			publishHomeassistantDiscoveryMessage(cfg, mc, dev.Name(), reg)
		}
	}
}

func homeassistantDiscoveryPeriodicModeRoutine(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	filter dataflow.RegisterFilterFunc,
) {
	interval := cfg.HomeassistantDiscovery().Interval()

	if cfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->homeassistantDiscovery: start periodic mode, send every %s",
			dev.Name(), mc.Name(), interval,
		)

		defer log.Printf(
			"device[%s]->mqttClient[%s]->homeassistantDiscovery: exit",
			dev.Name(), mc.Name(),
		)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, reg := range dev.RegisterDb().GetFiltered(filter) {
				publishHomeassistantDiscoveryMessage(cfg, mc, dev.Name(), reg)
			}
		}
	}
}

func publishHomeassistantDiscoveryMessage(
	cfg Config,
	mc mqttClient.Client,
	deviceName string,
	register dataflow.Register,
) {
	mCfg := cfg.HomeassistantDiscovery()

	var topic string
	var msg homeassistantDiscoveryMessage

	switch register.RegisterType() {
	case dataflow.NumberRegister:
		topic, msg = getHomeassistantDiscoverySensorMessage(
			cfg,
			deviceName,
			register,
			"{{ value_json.NumVal }}",
		)
	case dataflow.TextRegister:
		topic, msg = getHomeassistantDiscoverySensorMessage(
			cfg,
			deviceName,
			register,
			"{{ value_json.TextVal }}",
		)
	case dataflow.EnumRegister:
		// generate Jinja2 template to translate enumIdx to string
		enum := register.Enum()
		var valueTemplate strings.Builder
		op := "if"
		for idx, value := range enum {
			fmt.Fprintf(&valueTemplate, "{%% %s value_json.EnumIdx == %d %%}%s", op, idx, value)
			op = "elif"
		}
		valueTemplate.WriteString("{% endif %}")

		topic, msg = getHomeassistantDiscoverySensorMessage(
			cfg,
			deviceName,
			register,
			valueTemplate.String(),
		)
	default:
		return
	}

	log.Printf("homeassistantDiscovery[%s]: send %s %#v", mc.Name(), topic, msg)

	if payload, err := json.Marshal(msg); err != nil {
		log.Printf("homeassistantDiscovery: cannot generate discovery message: %s", err)
	} else {
		mc.Publish(
			topic,
			payload,
			mCfg.Qos(),
			mCfg.Retain(),
		)
	}
}

func getHomeassistantDiscoverySensorMessage(
	cfg Config,
	deviceName string,
	register dataflow.Register,
	valueTemplate string,
) (topic string, msg homeassistantDiscoveryMessage) {
	uniqueId := fmt.Sprintf("%s-%s", deviceName, CamelToSnakeCase(register.Name()))
	name := fmt.Sprintf("%s %s", deviceName, register.Description())

	topic = cfg.HomeassistantDiscoveryTopic("sensor", cfg.ClientId(), uniqueId)

	msg = homeassistantDiscoveryMessage{
		UniqueId:          uniqueId,
		Name:              name,
		StateTopic:        cfg.RealtimeTopic(deviceName, register.Name()),
		Availability:      getHomeassistantDiscoveryAvailabilityTopics(cfg, deviceName),
		AvailabilityMode:  "all",
		ValueTemplate:     valueTemplate,
		UnitOfMeasurement: register.Unit(),
	}

	return
}

func getHomeassistantDiscoveryAvailabilityTopics(cfg Config, deviceName string) (ret []homeassistantDiscoveryAvailabilityStruct) {
	if cfg.AvailabilityClient().Enabled() {
		ret = append(ret, homeassistantDiscoveryAvailabilityStruct{cfg.AvailabilityClientTopic()})
	}
	if cfg.AvailabilityDevice().Enabled() {
		ret = append(ret, homeassistantDiscoveryAvailabilityStruct{cfg.AvailabilityDeviceTopic(deviceName)})
	}

	return
}
