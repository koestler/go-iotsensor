package config

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	InvalidSyntaxConfig = `- -`

	InvalidUnknownVersionConfig = `
Version: 42
`

	ValidComplexConfig = `
Version: 1                                                 # configuration file format; must be set to 1 for >v2 of this tool.
ProjectTitle: Configurable Title of Project                # optional, default go-iotdevice: is shown in the http frontend
LogConfig: true                                            # optional, default False, outputs the used configuration including defaults on startup
LogWorkerStart: true                                       # optional, default False, outputs what devices and mqtt clients are started
LogStorageDebug: true                                      # optional, default False, outputs all write to the internal value storage

HttpServer:                                                # optional, when missing: http server is not started
  Bind: ::1                                                # optional, default ::1 (ipv6 loopback), what address to bind to, use "0:0:0:0" when started within docker
  Port: 8000                                               # optional, default 8000
  LogRequests: false                                       # optional, default true, enables the http access log to stdout
  # configure FrontendProxy xor FrontendPath
  FrontendProxy: "http://127.0.0.1:3000/"                  # optional, default deactivated; proxies the frontend to another server; useful for development
  FrontendPath: ./frontend-build/                          # optional, default "frontend-build": path to a static frontend build
  FrontendExpires: 1m                                      # optional, default 5min, what cache-control header to send for static frontend files
  ConfigExpires: 2m                                        # optional, default 1min, what cache-control header to send for configuration endpoints
  LogDebug: true                                          # optional, default false, output debug messages related to the http server

Modbus:                                                    # optional, when empty, no modbus handler is started
  bus0:                                                    # mandatory, an arbitrary name used for logging and for referencing in other config sections
    Device: /dev/ttyACM0                                   # mandatory, the RS485 serial device
    BaudRate: 9600                                         # mandatory, eg. 9600
    ReadTimeout: 100ms                                     # optional, default 100ms, how long to wait for a response
    LogDebug: true                                         # optional, default false, verbose debug log

Authentication:                                            # optional, when missing: login is disabled
  JwtSecret: 'aiziax9Hied0ier9Yo0Lo6bi3xahth7o'            # optional, default random, used to sign the JWT tokens
  JwtValidityPeriod: 2h                                    # optional, default 1h, users are logged out after this time
  HtaccessFile: ./my-auth.passwd                           # mandatory, where the file generated by htpasswd can be found

MqttClients:                                               # optional, when empty, no mqtt connection is made
  0-local:                                                 # mandatory, an arbitrary name used for logging and for referencing in other config sections
    Broker: tcp://mqtt.example.com:1883                    # mandatory, the URL to the server, use tcp:// or ssl://
    ProtocolVersion: 5                                     # optional, default 5, must be 5 always, only mqtt protocol version 5 is supported
    User: dev                                              # optional, default empty, the user used for authentication
    Password: zee4AhRi                                     # optional, default empty, the password used for authentication
    ClientId: server42                                     # optional, default go-iotdevice-UUID, mqtt client id, make sure it is unique per mqtt-server
    Qos: 0                                                 # optional, default 1, what quality-of-service level shall be used for published messages and subscriptions
    KeepAlive: 2m                                          # optional, default 60s, how often a ping is sent to keep the connection alive
    ConnectRetryDelay: 20s                                 # optional, default 10s, when disconnected: after what delay shall a connection attempt is made
    ConnectTimeout: 10s                                    # optional, default 5s, how long to wait for the SYN+ACK packet, increase on slow networks
    AvailabilityTopic: '%Prefix%/%ClientId%/status'        # optional, what topic to use for online/offline messages
    TelemetryInterval: 20s                                 # optional, default 10s, how often to sent telemetry mqtt messages, 0s disables tlemetry messages
    TelemetryTopic: '%Prefix%tele/%DeviceName%/state'      # optional, what topic to use for telemetry messages
    TelemetryRetain: true                                  # optional, default false, the mqtt retain flag for telemetry messages
    RealtimeEnable: true                                   # optional, default false, whether to enable sending realtime messages
    RealtimeTopic: '%Prefix%stat/%DeviceName%/%ValueName%' # optional, what topic to use for realtime messages
    RealtimeRetain: false                                  # optional, default true, the mqtt retain flag for realtime messages
    TopicPrefix: my-prefix                                 # optional, default empty, %Prefix% is replaced with this string
    LogDebug: true                                         # optional, default false, very verbose debug log of the mqtt connection
    LogMessages: true                                      # optional, default false, log all incoming mqtt messages
  1-remote:
    Broker: "ssl://eu1.cloud.thethings.network:8883"

HassDiscovery:                                             # optional, default, empty, defines which registers should be advertised via the homeassistant discovery mechanism
                                                           # You can have multiple sections to advertise on different topics, on different MqttServers of matching different registers
                                                           # each register is only advertised once per server / topic even if multiple entries match
  - TopicPrefix:                                           # optional, default 'homeassistant', the mqtt topic used for the discovery messages
    ViaMattClients:                                        # optional, default all clients, on what mqtt servers shall the registers by advertised
    Devices:                                               # optional, default all, a list of regular expressions against which devices names are matched (eg. "device1" or user ".*" for all devices)
      - bmv1                                               # use device identifiers of the VictronDevices, ModbusDevices etc. sections
    Categories:                                            # optional, default all, a list of regular expressions against which devices names are matched (eg. "device1" or user ".*" for all devices)
      - .*                                                 # match all categories; see the category field in /api/v2/views/dev/devices/DEVICE-NAME/registers
    Registers:                                             # optional, default all, a list of regular expressions against which register names are matched
      - Voltage$                                           # matches all registers with a name ending in Voltage, eg. MainVoltage, AuxVoltage
      - ˆBattery                                           # matches all registers with a name begining with Battery, eg. BatteryTemperature

VictronDevices:                                            # optional, a list of Victron Energy devices to connect to
  bmv0:                                                    # mandatory, an arbitrary name used for logging and for referencing in other config sections
    General:                                               # optional, this section is exactly the same for all devices
      SkipFields:                                          # optional, default empty, a list of field names that shall be ignored for this device
        - Temperature                                      # for BMV devices without a temperature sensor connect
        - AuxVoltage                                       # for BMV devices without a mid- or starter-voltage reading
      SkipCategories:                                      # optional, default empty, a list of category names that shall be ignored for this device
        - Settings                                         # for solar devices it might make sense to not fetch / output the settings
      TelemetryViaMqttClients:                             # optional, default all clients, to what mqtt servers shall telemetry messages be sent to
        - 0-local                                          # state the arbitrary name of the mqtt client as defined in the MqttClients section of this file
      RealtimeViaMqttClients:                              # optional, default all clients, to what mqtt servers shall realtime messages be sent to
        - 0-local
      RestartInterval: 200ms                               # optional, default 200ms, how fast to restart the device if it fails / disconnects
      RestartIntervalMaxBackoff: 1m                        # optional, default 1m; when it fails, the restart interval is exponentially increased up to this maximum
      LogDebug: false                                      # optional, default false, enable debug log output
      LogComDebug: false                                   # optional, default false, enable a verbose log of the communication with the device
    Device: /dev/serial/by-id/usb-VictronEnergy_BV_VE_Direct_cable_VEHTVQT-if00-port0 # mandatory except if Kind: Random*, the path to the usb-to-serial converter
    Kind: Vedirect                                         # mandatory, possibilities: Vedirect, RandomBmv, RandomSolar, always set to Vedirect expect for development

ModbusDevices:                                             # optional, a list of devices connected via ModBus
  modbus-rtu0:                                             # mandatory, an arbitrary name used for logging and for referencing in other config sections
    General:                                               # optional, this section is exactly the same for all devices
      SkipFields:                                          # optional, default empty, a list of field names that shall be ignored for this device
      SkipCategories:                                      # optional, default empty, a list of category names that shall be ignored for this device
      TelemetryViaMqttClients:                             # optional, default all clients, to what mqtt servers shall telemetry messages be sent to
      RealtimeViaMqttClients:                              # optional, default all clients, to what mqtt servers shall realtime messages be sent to
      RestartInterval: 200ms                               # optional, default 200ms, how fast to restart the device if it fails / disconnects
      RestartIntervalMaxBackoff: 1m                        # optional, default 1m; when it fails, the restart interval is exponentially increased up to this maximum
      LogDebug: false                                      # optional, default false, enable debug log output
      LogComDebug: false                                   # optional, default false, enable a verbose log of the communication with the device
    Bus: bus0                                              # mandatory, the identifier of the modbus to use
    Kind: WaveshareRtuRelay8                               # mandatory, type/model of the device; possibilities: WaveshareRtuRelay8
    Address: 0x01                                          # mandatory, the modbus address of the device in hex as a string, e.g. 0x0A
    Relays:                                                # optional: a map of custom labels for the relays
      CH1:
        Description: Lamp                                  # optional: show the CH1 relay as "Lamp" in the frontend
        OpenLabel: Off                                     # optional, default "open", a label for the open state
        ClosedLabel: On                                    # optional, default "closed", a label for the closed state
    PollInterval: 1s                                       # optional, default 1s, how often to fetch the device status

HttpDevices:                                               # optional, a list of devices controlled via http
  tcw241:                                                  # mandatory, an arbitrary name used for logging and for referencing in other config sections
    General:                                               # optional, this section is exactly the same for all devices
      SkipFields:                                          # optional, default empty, a list of field names that shall be ignored for this device
      SkipCategories:                                      # optional, default empty, a list of category names that shall be ignored for this device
      TelemetryViaMqttClients:                             # optional, default all clients, to what mqtt servers shall telemetry messages be sent to
      RealtimeViaMqttClients:                              # optional, default all clients, to what mqtt servers shall realtime messages be sent to
      RestartInterval: 200ms                               # optional, default 200ms, how fast to restart the device if it fails / disconnects
      RestartIntervalMaxBackoff: 1m                        # optional, default 1m; when it fails, the restart interval is exponentially increased up to this maximum
      LogDebug: false                                      # optional, default false, enable debug log output
      LogComDebug: false                                   # optional, default false, enable a verbose log of the communication with the device
    Url: http://control0/                                  # mandatory, URL to the device; supported protocol is http/https; e.g. http://device0.local/
    Kind: Teracom                                          # mandatory, type/model of the device; possibilities: Teracom, Shelly3m
    Username: admin                                        # optional, username used to log in
    Password: my-secret                                    # optional, password used to log in
    PollInterval: 1s                                       # optional, default 1s, how often to fetch the device status

MqttDevices:                                               # optional, a list of devices receiving its values via a mqtt server from another instance
  bmv1:                                                    # mandatory, an arbitrary name used for logging and for referencing in other config sections
    General:                                               # optional, this section is exactly the same for all devices
      SkipFields:                                          # optional, default empty, a list of field names that shall be ignored for this device
      SkipCategories:                                      # optional, default empty, a list of category names that shall be ignored for this device
      TelemetryViaMqttClients:                             # optional, default all clients, to what mqtt servers shall telemetry messages be sent to
      RealtimeViaMqttClients:                              # optional, default all clients, to what mqtt servers shall realtime messages be sent to
      RestartInterval: 200ms                               # optional, default 200ms, how fast to restart the device if it fails / disconnects
      RestartIntervalMaxBackoff: 1m                        # optional, default 1m; when it fails, the restart interval is exponentially increased up to this maximum
      LogDebug: false                                      # optional, default false, enable debug log output
      LogComDebug: false                                   # optional, default false, enable a verbose log of the communication with the device
    MqttTopics:                                            # mandatory, at least 1 topic must be defined
      - stat/go-iotdevice/bmv1/+                           # what topic to subscribe to; must match RealtimeTopic of the sending device; %ValueName% must be replaced by +
    MqttClients:                                           # optional, default all clients, on which mqtt server(s) we subscribe
      - 0-local                                            # identifier as defined in the MqttClients section

Views:                                                     # optional, a list of views (=categories in the frontend / paths in the api URLs)
  - Name: victron                                          # mandatory, a technical name used in the URLs
    Title: Victron                                         # mandatory, a nice title displayed in the frontend
    Devices:                                               # mandatory, a list of devices using
      - Name: bmv0                                         # mandatory, the arbitrary names defined above
        Title: Battery Monitor                             # mandatory, a nice title displayed in the frontend
        SkipFields:                                        # optional, default empty, field names that are omitted when displaying this view
        SkipCategories:                                    # optional, default empty, category names that are omitted when displaying this view
      - Name: modbus-rtu0                                  # mandatory, the arbitrary names defined above
        Title: Relay Board                                 # mandatory, a nice title displayed in the frontend
    Autoplay: true                                         # optional, default true, when true, live updates are enabled automatically when the view is open in the frontend
    AllowedUsers:                                          # optional, if empty, all users of the HtaccessFile are considered valid, otherwise only those listed here
      - test0                                              # username which is allowed to access this view
    Hidden: false                                          # optional, default false, if true, this view is not shown in the menu unless the user is logged in
`
)

func containsError(needle string, err []error) bool {
	for _, e := range err {
		if strings.Contains(e.Error(), needle) {
			return true
		}
	}
	return false
}

func TestReadConfig_InvalidSyntax(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidSyntaxConfig), true)
	if len(err) != 1 {
		t.Error("expect one error for invalid file")
	}
}

func TestReadConfig_NoVersion(t *testing.T) {
	_, err := ReadConfig([]byte(""), true)

	if !containsError("version must be defined", err) {
		t.Error("no version given; expect 'version must be defined'")
	}
}

func TestReadConfig_InvalidUnknownVersion(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidUnknownVersionConfig), true)
	if len(err) != 1 || err[0].Error() != "version=42 is not supported" {
		t.Errorf("expect 1 error: 'version=42 is not supported' but got: %v", err)
	}
}

// check that a complex example setting all available options is correctly read
func TestReadConfig_Complex(t *testing.T) {
	config, err := ReadConfig([]byte(ValidComplexConfig), true)
	if len(err) > 0 {
		t.Errorf("did not expect any errors, got: %v", err)
	}

	t.Logf("config=%v", config)

	// General Section
	if expected, got := 1, config.Version(); expected != got {
		t.Errorf("expect Version to be %d but got %d", expected, got)
	}

	if expected, got := "Configurable Title of Project", config.ProjectTitle(); expected != got {
		t.Errorf("expected ProjectTitle to be '%s but got '%s'", expected, got)
	}

	if !config.LogConfig() {
		t.Errorf("expect LogConfig to be True as configured")
	}

	if !config.LogWorkerStart() {
		t.Errorf("expect LogWorkerStart to be True as configured")
	}

	if !config.LogStorageDebug() {
		t.Errorf("expect LogStorageDebug to be True as configured")
	}

	{
		hs := config.HttpServer()
		if !hs.Enabled() {
			t.Error("expect HttpServer->Enabled to be True")
		}

		if expected, got := "::1", hs.Bind(); expected != got {
			t.Errorf("expect HttpServer->Bind to be '%s' but got '%s'", expected, got)
		}

		if expected, got := 8000, hs.Port(); expected != got {
			t.Errorf("expect HttpServer->Port to be %d but got %d", expected, got)
		}

		if hs.LogRequests() {
			t.Error("expect HttpServer->LogRequests to be False")
		}

		if expected, got := "http://127.0.0.1:3000/", hs.FrontendProxy().String(); expected != got {
			t.Errorf("expected HttpServer->FrontendProxy to be '%s' but got '%s'", expected, got)
		}

		if expected, got := "./frontend-build/", hs.FrontendPath(); expected != got {
			t.Errorf("expected HttpServer->FrontendPath to be '%s' but got '%s'", expected, got)
		}

		if expected, got := time.Minute, hs.FrontendExpires(); expected != got {
			t.Errorf("expected HttpServer->FrontendExpires to be %s but got %s", expected, got)
		}

		if expected, got := 2*time.Minute, hs.ConfigExpires(); expected != got {
			t.Errorf("expected HttpServer->ConfigExpires to be %s but got %s", expected, got)
		}

		if !hs.LogDebug() {
			t.Error("expect HttpServer->LogDebug to be True")
		}
	}

	{
		a := config.Authentication()

		if !a.Enabled() {
			t.Error("expect Authentication->Enabled to be True")
		}

		if expected, got := "aiziax9Hied0ier9Yo0Lo6bi3xahth7o", string(a.JwtSecret()); expected != got {
			t.Errorf("expect Authentication->JwtSecret to be '%s' but got '%s'", expected, got)
		}

		if expected, got := 2*time.Hour, a.JwtValidityPeriod(); expected != got {
			t.Errorf("expected Authentication->JwtValidityPeriod to be %s but got %s", expected, got)
		}

		if expected, got := "./my-auth.passwd", a.HtaccessFile(); expected != got {
			t.Errorf("expected Authentication->HtaccessFile to be '%s' but got '%s'", expected, got)
		}
	}

	if len(config.MqttClients()) != 2 {
		t.Error("expect len(config.MqttClients) == 2")
	}

	{
		mc := config.MqttClients()[0]

		if expected, got := "0-local", mc.Name(); expected != got {
			t.Errorf("expect Name of first MqttClient to be '%s' but got '%s'", expected, got)
		}

		if expected, got := "tcp://mqtt.example.com:1883", mc.Broker().String(); expected != got {
			t.Errorf("expect MqttClients->local->Broker to be '%s' but got '%s'", expected, got)
		}

		if expected, got := 5, mc.ProtocolVersion(); expected != got {
			t.Errorf("expect MqttClients->local->ProtocolVersion to be %d but got %d", expected, got)
		}

		if expected, got := "dev", mc.User(); expected != got {
			t.Errorf("expect MqttClients->local->User to be '%s' but got '%s'", expected, got)
		}

		if expected, got := "zee4AhRi", mc.Password(); expected != got {
			t.Errorf("expect MqttClients->local->Password to be '%s' but got '%s'", expected, got)
		}

		if expected, got := "server42", mc.ClientId(); expected != got {
			t.Errorf("expect MqttClients->local->ClientId to be '%s' but got '%s'", expected, got)
		}

		if expected, got := byte(0), mc.Qos(); expected != got {
			t.Errorf("expect MqttClients->local->Qos to be %d but got %d", expected, got)
		}

		if expected, got := 2*time.Minute, mc.KeepAlive(); expected != got {
			t.Errorf("expect MqttClients->local->KeepAlive to be '%s' but got '%s'", expected, got)
		}

		if expected, got := 20*time.Second, mc.ConnectRetryDelay(); expected != got {
			t.Errorf("expect MqttClients->local->ConnectRetryDelay to be '%s' but got '%s'", expected, got)
		}

		if expected, got := 10*time.Second, mc.ConnectTimeout(); expected != got {
			t.Errorf("expect MqttClients->local->ConnectTimeout to be '%s' but got '%s'", expected, got)
		}

		if expected, got := "%Prefix%/%ClientId%/status", mc.AvailabilityTopic(); expected != got {
			t.Errorf("expect MqttClients->local->AvailabilityTopic to be '%s' but got '%s'", expected, got)
		}

		if expected, got := 20*time.Second, mc.TelemetryInterval(); expected != got {
			t.Errorf("expect MqttClients->local->TelemetryInterval to be '%s' but got '%s'", expected, got)
		}

		if expected, got := "%Prefix%tele/%DeviceName%/state", mc.TelemetryTopic(); expected != got {
			t.Errorf("expect MqttClients->local->TelemetryTopic to be '%s' but got '%s'", expected, got)
		}

		if !mc.TelemetryRetain() {
			t.Error("expect MqttClients->local->TelemetryRetain to be true")
		}

		if !mc.RealtimeEnable() {
			t.Error("expect MqttClients->local->RealtimeEnable to be true")
		}

		if expected, got := "%Prefix%stat/%DeviceName%/%ValueName%", mc.RealtimeTopic(); expected != got {
			t.Errorf("expect MqttClients->local->RealtimeTopic to be '%s' but got '%s'", expected, got)
		}

		if mc.RealtimeRetain() {
			t.Error("expect MqttClients->RealtimeRetain to be false")
		}

		if expected, got := "my-prefix", mc.TopicPrefix(); expected != got {
			t.Errorf("expect MqttClients->local->TopicPrefix to be '%s' but got '%s'", expected, got)
		}

		if !mc.LogDebug() {
			t.Error("expect MqttClients->local->LogDebug to be True")
		}

		if !mc.LogMessages() {
			t.Error("expect MqttClients->local->LogMessages to be True")
		}
	}

	{
		mc := config.MqttClients()[1]

		if expected, got := "1-remote", mc.Name(); expected != got {
			t.Errorf("expect Name of second MqttClient to be '%s' but got '%s'", expected, got)
		}

		if expected, got := "ssl://eu1.cloud.thethings.network:8883", mc.Broker().String(); expected != got {
			t.Errorf("expect MqttClients->local->Broker to be '%s' but got '%s'", expected, got)
		}

		if expected, got := 5, mc.ProtocolVersion(); expected != got {
			t.Errorf("expect MqttClients->local->ProtocolVersion to be %d but got %d", expected, got)
		}

		if expected, got := "", mc.User(); expected != got {
			t.Errorf("expect MqttClients->local->User to be '%s' but got '%s'", expected, got)
		}
	}

	// test config output does not crash
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	if err := config.PrintConfig(); err != nil {
		t.Errorf("expected no error. Got: %s", err)
	}
	t.Log(buf.String())
}

// check that configuration file in the documentation do not contain any errors
func TestReadConfig_DocumentationFullConfig(t *testing.T) {
	_, err := ReadConfigFile("", "../documentation/full-config.yaml", true)
	if len(err) > 0 {
		t.Errorf("did not expect any error, got %v", err)
	}
}
