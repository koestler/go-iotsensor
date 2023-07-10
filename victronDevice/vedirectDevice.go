package victronDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/vedirect"
	"log"
	"strings"
	"time"
)

func runVedirect(ctx context.Context, c *DeviceStruct, output dataflow.Fillable) (err error, immediateError bool) {
	log.Printf("device[%s]: start vedirect source", c.deviceConfig.Name())

	// open vedirect device
	vd, err := vedirect.Open(c.victronConfig.Device(), c.deviceConfig.LogComDebug())
	if err != nil {
		return err, true
	}
	defer func() {
		if err := vd.Close(); err != nil {
			log.Printf("device[%s]: vd.Close failed: %s", c.deviceConfig.Name(), err)
		}
	}()

	// send ping
	if err := vd.VeCommandPing(); err != nil {
		return fmt.Errorf("ping failed: %s", err), true
	}

	// send connected now, disconnected when this routine stops
	device.SendConnteced(c.Config().Name(), output)
	defer func() {
		device.SendDisconnected(c.Config().Name(), output)
	}()

	// get deviceId
	deviceId, err := vd.VeCommandDeviceId()
	if err != nil {
		return fmt.Errorf("cannot get DeviceId: %s", err), true
	}

	deviceString := deviceId.String()
	if len(deviceString) < 1 {
		return fmt.Errorf("unknown deviceId=%x", err), true
	}

	log.Printf("device[%s]: source: connect to %s", c.deviceConfig.Name(), deviceString)
	c.model = deviceString

	// get relevant registers
	{
		registers := RegisterFactoryByProduct(deviceId)
		if registers == nil {
			return fmt.Errorf("no registers found for deviceId=%x", deviceId), true
		}
		// filter registers by skip list
		c.registers = FilterRegisters(registers, c.deviceConfig.SkipFields(), c.deviceConfig.SkipCategories())
	}

	// start polling loop
	fetchStaticCounter := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, false
		case <-ticker.C:
			start := time.Now()

			// flush async data
			vd.RecvFlush()

			// execute a Ping at the beginning and after each error
			pingNeeded := true

			for _, register := range c.registers {
				// only fetch static registers seldomly
				if register.Static() && (fetchStaticCounter%60 != 0) {
					continue
				}

				if pingNeeded {
					if err := vd.VeCommandPing(); err != nil {
						return fmt.Errorf("device[%s]: source: VeCommandPing failed: %s", c.deviceConfig.Name(), err), false
					}
				}

				switch register.RegisterType() {
				case dataflow.NumberRegister:
					var value float64
					if register.Signed() {
						var intValue int64
						intValue, err = vd.VeCommandGetInt(register.Address())
						value = float64(intValue)
					} else {
						var intValue uint64
						intValue, err = vd.VeCommandGetUint(register.Address())
						value = float64(intValue)
					}

					if err != nil {
						log.Printf("device[%s]: fetching number register failed: %v", c.deviceConfig.Name(), err)
					} else {
						output.Fill(dataflow.NewNumericRegisterValue(
							c.deviceConfig.Name(),
							register,
							value/float64(register.Factor())+register.Offset(),
						))
					}
				case dataflow.TextRegister:
					value, err := vd.VeCommandGetString(register.Address())

					if err != nil {
						log.Printf("device[%s]: fetching text register failed: %v", c.deviceConfig.Name(), err)
					} else {
						output.Fill(dataflow.NewTextRegisterValue(
							c.deviceConfig.Name(),
							register,
							strings.TrimSpace(value),
						))
					}
				case dataflow.EnumRegister:
					var intValue uint64
					intValue, err = vd.VeCommandGetUint(register.Address())

					if err != nil {
						log.Printf("device[%s]: fetching enum register failed: %v", c.deviceConfig.Name(), err)
					} else {
						output.Fill(dataflow.NewEnumRegisterValue(
							c.deviceConfig.Name(),
							register,
							int(intValue),
						))
					}
				}

				pingNeeded = err != nil
			}

			c.SetLastUpdatedNow()

			fetchStaticCounter++

			if c.deviceConfig.LogDebug() {
				log.Printf(
					"device[%s]: registers fetched, took=%.3fs",
					c.deviceConfig.Name(),
					time.Since(start).Seconds(),
				)
			}
		}
	}
}
