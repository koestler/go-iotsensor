package vedata

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/koestler/go-ve-sensor/bmv"
	"log"
	"time"
)

type DeviceId uint64

type Device struct {
	Name          string
	LastUpdate    time.Time
	NumericValues bmv.NumericValues
}

type readDeviceOp struct {
	deviceId DeviceId
	response chan bool
	err      error
	device   Device
}

type readDeviceIdsOp struct {
	response chan []DeviceId
}

type writeOp struct {
	deviceId      DeviceId
	numericValues bmv.NumericValues
	response      chan bool
}

type DbType map[DeviceId]*Device

var running bool
var db DbType
var readDeviceChan chan *readDeviceOp
var readDeviceIdsChan chan *readDeviceIdsOp
var writes chan *writeOp

func init() {
	running = false
	db = make(map[DeviceId]*Device)

	readDeviceChan = make(chan *readDeviceOp)
	readDeviceIdsChan = make(chan *readDeviceIdsOp)
	writes = make(chan *writeOp)
}

func CreateDevice(name string) (deviceId DeviceId) {
	if running {
		log.Panic("must no call vedata.CreateDevice after vedata.Run")
	}

	deviceId = DeviceId(len(db) + 1)

	db[deviceId] = &Device{
		Name:          name,
		NumericValues: make(bmv.NumericValues),
	}

	return
}

func (deviceId DeviceId) ReadDevice() (device Device, err error) {
	read := &readDeviceOp{
		deviceId: deviceId,
		response: make(chan bool),
		err:      nil,
		device:   Device{},
	}
	readDeviceChan <- read
	<-read.response

	return read.device, read.err
}

func ReadDeviceIds() (ret []DeviceId) {
	read := &readDeviceIdsOp{
		response: make(chan []DeviceId)}
	readDeviceIdsChan <- read
	ret = <-read.response
	return
}

func (deviceId DeviceId) Write(numericValues bmv.NumericValues) {
	write := &writeOp{
		deviceId:      deviceId,
		numericValues: numericValues,
		response:      make(chan bool),
	}
	writes <- write
	<-write.response
}

func clone(a, b interface{}) {
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	dec := gob.NewDecoder(buff)
	enc.Encode(a)
	dec.Decode(b)
}

func Run() {
	go func() {
		running = true
		for {
			select {
			case write := <-writes:
				device, ok := db[write.deviceId]
				if ok {
					for k, v := range write.numericValues {
						device.NumericValues[k] = v
					}
					device.LastUpdate = time.Now()
				}
				write.response <- true
			case read := <-readDeviceChan:
				device, ok := db[read.deviceId]
				if !ok {
					read.err = errors.New("device not found")
				} else {
					// make deep copy
					clone(device, &read.device)
				}

				read.response <- true
			case read := <-readDeviceIdsChan:
				deviceIds := make([]DeviceId, len(db))
				i := 0
				for k, _ := range db {
					deviceIds[i] = k
					i++
				}
				read.response <- deviceIds
			}
		}
	}()
}