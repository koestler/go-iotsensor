package main
/*
import (
	"github.com/koestler/go-ve-sensor/vedevices"
	"github.com/koestler/go-ve-sensor/config"
	"github.com/koestler/go-ve-sensor/vedata"
	"github.com/koestler/go-ve-sensor/vedirect"
	"log"
	"time"
)

func BmvStart(config config.BmvConfig) {

	// create new db device connection
	bmvDeviceId := vedata.CreateDevice(config)

	// create
	vd, err := vedirect.Open(config.Device)
	if err != nil {
		log.Fatalf("device: cannot create vedirect, device=%v", config.Device)
		return
	}

	// start vedevices reader
	go func() {
		numericValues := make(vedevices.NumericValues)

		for _ = range time.Tick(400 * time.Millisecond) {
			if err := vd.VeCommandPing(); err != nil {
				log.Printf("device: VeCommandPing failed: %v", err)
			}

			var registers vedevices.Registers

			switch config.Model {
			case "bmv700":
				registers = vedevices.RegisterListBmv700
				break
			case "bmv702":
				registers = vedevices.RegisterListBmv702
				break
			default:
				log.Fatalf("device: unknown Bmv.Model: %v", config.Model)
			}

			for regName, reg := range registers {
				if numericValue, err := reg.RecvNumeric(vd); err != nil {
					log.Printf("device: vedevices.RecvNumeric failed: %v", err)
				} else {
					numericValues[regName] = numericValue
					if config.DebugPrint {
						log.Printf("%v : %v = %v", config.Name, regName, numericValue)
					}
				}
			}

			bmvDeviceId.Write(numericValues)
		}
	}()

}

*/