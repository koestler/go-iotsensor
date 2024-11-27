package generator_test

import (
	"fmt"
	"github.com/koestler/go-iotdevice/v3/generator"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestController(t *testing.T) {
	params := generator.Params{
		PrimingTimeout:           10 * time.Second,
		CrankingTimeout:          20 * time.Second,
		WarmUpTimeout:            10 * time.Minute,
		WarmUpTemp:               40,
		EngineCoolDownTimeout:    5 * time.Minute,
		EngineCoolDownTemp:       60,
		EnclosureCoolDownTimeout: 15 * time.Minute,
		EnclosureCoolDownTemp:    50,

		// IO Check
		EngineTempMin: 0,
		EngineTempMax: 100,
		AuxTemp0Min:   0,
		AuxTemp0Max:   100,
		AuxTemp1Min:   0,
		AuxTemp1Max:   100,

		// Output Check
		UMin:    210,
		UMax:    250,
		FMin:    45,
		FMax:    55,
		PMax:    1000,
		PTotMax: 2000,
	}

	t0, _ := time.Parse(time.RFC3339, "2000-01-01T00:00:00Z")
	initialInputs := generator.Inputs{Time: t0}

	t.Run("simpleSuccessfulRun", func(t *testing.T) {
		c := generator.NewController(params, generator.Off, initialInputs)

		stateTracker := newTracker[generator.State](t)
		c.OnStateUpdate = stateTracker.OnUpdateFunc()
		outputTracker := newTracker[generator.Outputs](t)
		c.OnOutputUpdate = outputTracker.OnUpdateFunc()

		c.Run()
		defer c.End()

		setInp := func(f func(i generator.Inputs) generator.Inputs) {
			c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
				i = f(i)
				t.Logf("inputs: %v", i)
				return i
			})
		}

		// initial state
		stateTracker.AssertLatest(t, generator.State{Node: generator.Off, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{})

		// go to ready
		setInp(func(i generator.Inputs) generator.Inputs {
			i.IOAvailable = true
			i.EngineTemp = 20
			i.AuxTemp0 = 20
			i.AuxTemp1 = 20
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Ready, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{IoCheck: true})

		// go to priming
		setInp(func(i generator.Inputs) generator.Inputs {
			i.ArmSwitch = true
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Ready, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{IoCheck: true})

		setInp(func(i generator.Inputs) generator.Inputs {
			i.CommandSwitch = true
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Priming, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, IoCheck: true})

		// stay in priming
		t1 := t0.Add(params.PrimingTimeout / 2)
		setInp(func(i generator.Inputs) generator.Inputs {
			i.Time = t1
			return i
		})

		// go to cranking
		t2 := t0.Add(params.PrimingTimeout)
		setInp(func(i generator.Inputs) generator.Inputs {
			i.Time = t2
			return i
		})

		stateTracker.AssertLatest(t, generator.State{Node: generator.Cranking, Changed: t2})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, Starter: true, IoCheck: true})

		// go to warm up
		setInp(func(i generator.Inputs) generator.Inputs {
			i.OutputAvailable = true
			i.U0 = 220
			i.U1 = 220
			i.U2 = 220
			i.F = 50
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.WarmUp, Changed: t2})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, IoCheck: true, OutputCheck: true})

		// go to producing
		setInp(func(i generator.Inputs) generator.Inputs {
			i.EngineTemp = 45
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Producing, Changed: t2})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, Load: true, IoCheck: true, OutputCheck: true})

		// running, engine getting warm, frequency fluctuating
		t3 := t2.Add(time.Second)
		setInp(func(i generator.Inputs) generator.Inputs {
			i.Time = t3
			i.EngineTemp = 70
			i.F = 48
			i.L0 = 1000
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Producing, Changed: t2})
		outputTracker.AssertLatest(t, generator.Outputs{
			Fan: true, Pump: true, Ignition: true, Load: true,
			TimeInState: time.Second, IoCheck: true, OutputCheck: true,
		})

		t4 := t3.Add(time.Second)
		setInp(func(i generator.Inputs) generator.Inputs {
			i.Time = t4
			i.EngineTemp = 72
			i.F = 51
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Producing, Changed: t2})
		outputTracker.AssertLatest(t, generator.Outputs{
			Fan: true, Pump: true, Ignition: true, Load: true,
			TimeInState: 2 * time.Second, IoCheck: true, OutputCheck: true,
		})

		// go to engine cool down
		t5 := t4.Add(time.Second)
		setInp(func(i generator.Inputs) generator.Inputs {
			i.Time = t5
			i.CommandSwitch = false
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.EngineCoolDown, Changed: t5})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, IoCheck: true, OutputCheck: true})

		// go to enclosure cool down
		t6 := t5.Add(time.Second)
		setInp(func(i generator.Inputs) generator.Inputs {
			i.Time = t6
			i.EngineTemp = 55
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.EnclosureCoolDown, Changed: t6})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, IoCheck: true, OutputCheck: true})

		// stay in enclosure cool down, engine has stopped
		t7 := t6.Add(time.Second)
		setInp(func(i generator.Inputs) generator.Inputs {
			i.Time = t7
			i.F = 0
			i.U0 = 10
			i.U1 = 10
			i.U2 = 10
			i.L0 = 2
			i.L1 = 2
			i.L2 = 2
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.EnclosureCoolDown, Changed: t6})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, IoCheck: true, TimeInState: time.Second})

		// go to ready
		t8 := t7.Add(time.Minute)
		setInp(func(i generator.Inputs) generator.Inputs {
			i.Time = t8
			i.EngineTemp = 45
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Ready, Changed: t8})
		outputTracker.AssertLatest(t, generator.Outputs{IoCheck: true})
	})
}

func BenchmarkUpdateInputs(b *testing.B) {
	b.Run("inputChanges", func(b *testing.B) {
		c := generator.NewController(generator.Params{}, generator.Off, generator.Inputs{})
		c.Run()
		defer c.End()

		for i := 0; i < b.N; i++ {
			c.UpdateInputs(func(i generator.Inputs) generator.Inputs {
				i.Time = time.Now()
				return i
			})
		}
	})
}

func BenchmarkUpdateInputsSync(b *testing.B) {
	b.Run("inputChanges", func(b *testing.B) {
		c := generator.NewController(generator.Params{}, generator.Off, generator.Inputs{})
		c.Run()
		defer c.End()

		for i := 0; i < b.N; i++ {
			c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
				i.Time = time.Now()
				return i
			})
		}
	})
}

func TestSyncWg(t *testing.T) {
	events := make([]string, 0, 20)
	eventsLock := sync.Mutex{}
	addEvent := func(pos string, v int) {
		eventsLock.Lock()
		defer eventsLock.Unlock()
		events = append(events, fmt.Sprintf("%02d: %v", v, pos))
	}

	type Change struct {
		v  int
		wg *sync.WaitGroup
	}

	inpChan := make(chan Change)
	oupChan := make(chan Change)

	go func() {
		for o := range oupChan {
			addEvent("C", o.v)
			o.wg.Done()
		}
	}()

	syncFunc := func(v int) {
		addEvent("A", v)
		wg := &sync.WaitGroup{}
		wg.Add(1)
		inpChan <- Change{
			v:  v,
			wg: wg,
		}
		wg.Wait()
		addEvent("D", v)
	}

	go func() {
		for c := range inpChan {
			addEvent("B", c.v)
			oupChan <- Change{
				v:  c.v,
				wg: c.wg,
			}
		}
	}()

	for i := 0; i < 20; i++ {
		syncFunc(i)
	}

	eventsLock.Lock()
	defer eventsLock.Unlock()

	sortedEvents := append([]string{}, events...)
	sort.Strings(sortedEvents)

	if !reflect.DeepEqual(events, sortedEvents) {
		t.Error("events not sorted")
		t.Logf("events: %s", strings.Join(events, "\n"))
		t.Logf("sortedEvents: %s", strings.Join(sortedEvents, "\n"))
	}
}
