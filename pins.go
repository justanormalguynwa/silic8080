package main

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// again, if i fuck up anything blame wikipedia cause like i get the thing from it idk
// pin state or we all getting deported
type PIN_STATE int

const (
	PIN_LOW  PIN_STATE = 0
	PIN_HIGH PIN_STATE = 1
	PIN_Z    PIN_STATE = 2
)

// i just watched a tutorial of how to construct a house with javascript and then had to map a chip pin by pin what the fuck
var PIN_MAP = map[string]int{
	"A0": 25, "A1": 26, "A2": 27, "A3": 29,
	"A4": 30, "A5": 31, "A6": 32, "A7": 33,
	"A8": 34, "A9": 35, "A10": 1, "A11": 40,
	"A12": 37, "A13": 38, "A14": 39, "A15": 36,
	"D0": 10, "D1": 9, "D2": 8, "D3": 7,
	"D4": 6, "D5": 5, "D6": 4, "D7": 3,
	"SYNC": 19, "DBIN": 17, "WR": 18, "READY": 23,
	"WAIT": 24, "HOLD": 13, "HLDA": 21, "INT": 14,
	"INTE": 16, "RESET": 12, "PHI1": 22, "PHI2": 15,
}

// state sponsored bits
const (
	STATE_INTA  = 0x01 // D0 - interruptsinterruptsinterruptsinterruptsinterruptsinterruptsinterruptsinterruptsinterruptsinterruptsinterrupts :fal:
	STATE_WO    = 0x02 // D1
	STATE_STACK = 0x04 // D2
	STATE_HLTA  = 0x08 // D3
	STATE_OUT   = 0x10 // D4
	STATE_M1    = 0x20 // D5
	STATE_INP   = 0x40 // D6
	STATE_MEMR  = 0x80 // D7
)

// timing
const (
	CLOCK_PERIOD_NS      = 500
	PHI1_DURATION_NS     = 240
	PHI2_DURATION_NS     = 240
	SETUP_TIME_NS        = 50
	HOLD_TIME_NS         = 30
	PROPAGATION_DELAY_NS = 20
)

// it do the emulation of the pin
type PIN_EMULATOR struct {
	PINS              [41]atomic.Value // index 0 unused
	ADDRESS_BUS       atomic.Value
	DATA_BUS          atomic.Value
	CONTROL_SIGNALS   sync.Map
	CONNECTED_DEVICES []DEVICE_INTERFACE
	DEVICE_MUTEX      sync.RWMutex
	// clock stuff
	CLOCK_PHI1    chan bool
	CLOCK_PHI2    chan bool
	CLOCK_RUNNING atomic.Bool
	CLOCK_FREQ_HZ atomic.Uint64
	CLOCK_CANCEL  context.CancelFunc
	CLOCK_CTX     context.Context
	// bus arbitration
	BUS_OWNER          atomic.Value
	BUS_REQUESTS       sync.Map
	PENDING_INTERRUPTS atomic.Uint32
	MACHINE_CYCLE      atomic.Uint32
	INSTRUCTION_CYCLE  atomic.Uint32
	PROCESSOR_STATE    atomic.Value
	WAIT_STATES        atomic.Uint32
	// timing
	TIMING_ENABLED  atomic.Bool
	LAST_TRANSITION atomic.Int64 // nanoseconds since epoch
}

// implement this for support chip idk
type DEVICE_INTERFACE interface {
	GET_DEVICE_ID() string
	ON_BUS_CYCLE(address uint16, data uint8, control_signals map[string]PIN_STATE) uint8
	REQUEST_BUS() bool
	RELEASE_BUS()
	IS_SELECTED(address uint16) bool
	RESET()
	GET_PRIORITY() int // for interrupt handling, lower number = higher priority
}

// i think NEW_PIN_EMULATOR creates a new pin emulator
func NEW_PIN_EMULATOR() *PIN_EMULATOR {
	ctx, cancel := context.WithCancel(context.Background())
	emulator := &PIN_EMULATOR{
		CONNECTED_DEVICES: make([]DEVICE_INTERFACE, 0),
		CLOCK_PHI1:        make(chan bool, 10),
		CLOCK_PHI2:        make(chan bool, 10),
		CLOCK_CANCEL:      cancel,
		CLOCK_CTX:         ctx,
	}
	for i := 1; i <= 40; i++ {
		emulator.PINS[i].Store(PIN_LOW)
	}
	emulator.ADDRESS_BUS.Store(uint16(0))
	emulator.DATA_BUS.Store(uint8(0))
	emulator.BUS_OWNER.Store("CPU")
	emulator.PROCESSOR_STATE.Store(uint8(0))
	emulator.CLOCK_FREQ_HZ.Store(uint64(2000000)) // 2mhz
	emulator.TIMING_ENABLED.Store(true)
	emulator.CLOCK_RUNNING.Store(false)
	go emulator.GENERATE_CLOCK()
	return emulator
}

//gimp
func (pe *PIN_EMULATOR) SET_PIN(pin_name string, state PIN_STATE) {
	if pin_num, exists := PIN_MAP[pin_name]; exists {
		if pe.TIMING_ENABLED.Load() {
			pe.SIMULATE_PROPAGATION_DELAY()
		}
		old_state := pe.PINS[pin_num].Load().(PIN_STATE)
		pe.PINS[pin_num].Store(state)
		if old_state != state {
			pe.HANDLE_PIN_CHANGE(pin_name, old_state, state)
		}
	}
}
func (pe *PIN_EMULATOR) GET_PIN(pin_name string) PIN_STATE {
	if pin_num, exists := PIN_MAP[pin_name]; exists {
		return pe.PINS[pin_num].Load().(PIN_STATE)
	}
	return PIN_LOW // default to low
}
func (pe *PIN_EMULATOR) HANDLE_PIN_CHANGE(pin_name string, old_state, new_state PIN_STATE) {
	switch pin_name {
	case "SYNC":
		if new_state == PIN_HIGH {
			pe.START_MACHINE_CYCLE()
		}
	case "DBIN":
		if new_state == PIN_HIGH {
			pe.HANDLE_READ_CYCLE()
		}
	case "WR":
		if old_state == PIN_HIGH && new_state == PIN_LOW { // falling edge trigger
			pe.HANDLE_WRITE_CYCLE()
		}
	case "READY":
		if new_state == PIN_LOW {
			pe.INSERT_WAIT_STATE()
		} else {
			pe.CLEAR_WAIT_STATES()
		}
	case "HOLD":
		if new_state == PIN_HIGH {
			pe.ACKNOWLEDGE_BUS_REQUEST()
		}
	case "INT":
		if new_state == PIN_HIGH {
			pe.HANDLE_INTERRUPT_REQUEST()
		}
	case "RESET":
		if new_state == PIN_LOW { // reset is
			pe.RESET_SYSTEM()
		}
	}
}
func (pe *PIN_EMULATOR) SET_ADDRESS_BUS(address uint16) {
	if pe.TIMING_ENABLED.Load() {
		pe.SIMULATE_SETUP_TIME()
	}
	pe.ADDRESS_BUS.Store(address)
	// set individual address pins
	for i := 0; i < 16; i++ {
		pin_name := MAP_ADDRESS_PIN_NAME(i)
		if (address>>i)&1 == 1 {
			pe.SET_PIN(pin_name, PIN_HIGH)
		} else {
			pe.SET_PIN(pin_name, PIN_LOW)
		}
	}
}
func (pe *PIN_EMULATOR) GET_ADDRESS_BUS() uint16 {
	return pe.ADDRESS_BUS.Load().(uint16)
}
func (pe *PIN_EMULATOR) SET_DATA_BUS(data uint8) {
	if pe.TIMING_ENABLED.Load() {
		pe.SIMULATE_SETUP_TIME()
	}
	pe.DATA_BUS.Store(data)
	// ihwrfweeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee888888888888888ewifnndskhjh
	for i := 0; i < 8; i++ {
		pin_name := MAP_DATA_PIN_NAME(i)
		if (data>>i)&1 == 1 {
			pe.SET_PIN(pin_name, PIN_HIGH)
		} else {
			pe.SET_PIN(pin_name, PIN_LOW)
		}
	}
}
func (pe *PIN_EMULATOR) GET_DATA_BUS() uint8 {
	// read from actual pins in case a device is driving the bus
	var data uint8 = 0
	for i := 0; i < 8; i++ {
		pin_name := MAP_DATA_PIN_NAME(i)
		if pe.GET_PIN(pin_name) == PIN_HIGH {
			data |= (1 << i)
		}
	}
	pe.DATA_BUS.Store(data)
	return data
}

// guys what does this mean
func MAP_ADDRESS_PIN_NAME(bit int) string {
	pin_names := [16]string{
		"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7",
		"A8", "A9", "A10", "A11", "A12", "A13", "A14", "A15",
	}
	if bit >= 0 && bit < 16 {
		return pin_names[bit]
	}
	return "INVALID"
}
func MAP_DATA_PIN_NAME(bit int) string {
	pin_names := [8]string{"D0", "D1", "D2", "D3", "D4", "D5", "D6", "D7"}
	if bit >= 0 && bit < 8 {
		return pin_names[bit]
	}
	return "INVALID"
}
func (pe *PIN_EMULATOR) CONNECT_DEVICE(device DEVICE_INTERFACE) {
	pe.DEVICE_MUTEX.Lock()
	defer pe.DEVICE_MUTEX.Unlock()
	priority := device.GET_PRIORITY()
	inserted := false
	for i, existing := range pe.CONNECTED_DEVICES {
		if priority < existing.GET_PRIORITY() {
			// insert here to maintain priority order
			pe.CONNECTED_DEVICES = append(pe.CONNECTED_DEVICES[:i],
				append([]DEVICE_INTERFACE{device}, pe.CONNECTED_DEVICES[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		pe.CONNECTED_DEVICES = append(pe.CONNECTED_DEVICES, device)
	}
}
func (pe *PIN_EMULATOR) DISCONNECT_DEVICE(device_id string) bool {
	pe.DEVICE_MUTEX.Lock()
	defer pe.DEVICE_MUTEX.Unlock()
	for i, device := range pe.CONNECTED_DEVICES {
		if device.GET_DEVICE_ID() == device_id {
			pe.CONNECTED_DEVICES = append(pe.CONNECTED_DEVICES[:i], pe.CONNECTED_DEVICES[i+1:]...)
			device.RELEASE_BUS() // make sure it's not holding the bus hostage
			return true
		}
	}
	return false
}
func (pe *PIN_EMULATOR) START_MACHINE_CYCLE() {
	cycle_count := pe.MACHINE_CYCLE.Add(1)
	processor_state := pe.GET_DATA_BUS()
	pe.PROCESSOR_STATE.Store(processor_state)
	// decode what kind of cycle this is
	pe.DECODE_MACHINE_CYCLE(processor_state)
	// broadcast to all devices because dictatorship
	current_address := pe.GET_ADDRESS_BUS()
	control_signals := pe.BUILD_CONTROL_SIGNAL_MAP()
	pe.DEVICE_MUTEX.RLock()
	for _, device := range pe.CONNECTED_DEVICES {
		// run device responses in parallel
		go func(d DEVICE_INTERFACE) {
			d.ON_BUS_CYCLE(current_address, processor_state, control_signals)
		}(device)
	}
	pe.DEVICE_MUTEX.RUnlock()
}
func (pe *PIN_EMULATOR) DECODE_MACHINE_CYCLE(state uint8) {
	// decode the processor state bits to understand what's happening
	if state&STATE_M1 != 0 {
		pe.INSTRUCTION_CYCLE.Add(1)
	}
	if state&STATE_INTA != 0 {
		pe.HANDLE_INTERRUPT_ACKNOWLEDGE()
	}
	if state&STATE_HLTA != 0 {
		pe.HANDLE_HALT_STATE()
	}
}
func (pe *PIN_EMULATOR) BUILD_CONTROL_SIGNAL_MAP() map[string]PIN_STATE {
	signals := make(map[string]PIN_STATE)
	// grab all the control signals
	control_pins := []string{
		"SYNC", "DBIN", "WR", "READY", "WAIT", "HOLD",
		"HLDA", "INT", "INTE", "RESET", "PHI1", "PHI2",
	}
	for _, pin := range control_pins {
		signals[pin] = pe.GET_PIN(pin)
	}
	return signals
}
func (pe *PIN_EMULATOR) HANDLE_READ_CYCLE() {
	address := pe.GET_ADDRESS_BUS()
	if pe.TIMING_ENABLED.Load() {
		pe.SIMULATE_ACCESS_TIME()
	}
	pe.DEVICE_MUTEX.RLock()
	defer pe.DEVICE_MUTEX.RUnlock()
	for _, device := range pe.CONNECTED_DEVICES {
		if device.IS_SELECTED(address) {
			// let the device drive the data bus
			control_signals := pe.BUILD_CONTROL_SIGNAL_MAP()
			data := device.ON_BUS_CYCLE(address, 0, control_signals)
			pe.SET_DATA_BUS(data)
			break // only one device should respond
		}
	}
}
func (pe *PIN_EMULATOR) HANDLE_WRITE_CYCLE() {
	address := pe.GET_ADDRESS_BUS()
	data := pe.GET_DATA_BUS()
	if pe.TIMING_ENABLED.Load() {
		pe.SIMULATE_ACCESS_TIME()
	}
	pe.DEVICE_MUTEX.RLock()
	defer pe.DEVICE_MUTEX.RUnlock()
	// broadcast write to all devices, let them figure out if they care
	control_signals := pe.BUILD_CONTROL_SIGNAL_MAP()
	for _, device := range pe.CONNECTED_DEVICES {
		if device.IS_SELECTED(address) {
			device.ON_BUS_CYCLE(address, data, control_signals)
			break // again, only one should respond
		}
	}
	if pe.TIMING_ENABLED.Load() {
		pe.SIMULATE_HOLD_TIME()
	}
}

// the thing do interuopt
func (pe *PIN_EMULATOR) HANDLE_INTERRUPT_REQUEST() {
	if pe.GET_PIN("INTE") == PIN_HIGH {
		pe.PENDING_INTERRUPTS.Add(1)
	}
}
func (pe *PIN_EMULATOR) HANDLE_INTERRUPT_ACKNOWLEDGE() {
	if pe.PENDING_INTERRUPTS.Load() > 0 {
		pe.PENDING_INTERRUPTS.Add(^uint32(0))
		pe.DEVICE_MUTEX.RLock()
		for _, device := range pe.CONNECTED_DEVICES {
			control_signals := pe.BUILD_CONTROL_SIGNAL_MAP()
			vector := device.ON_BUS_CYCLE(0, STATE_INTA, control_signals)
			if vector != 0 {
				pe.SET_DATA_BUS(vector)
				break
			}
		}
		pe.DEVICE_MUTEX.RUnlock()
	}
}

// wait state stuff fuck you
func (pe *PIN_EMULATOR) INSERT_WAIT_STATE() {
	pe.WAIT_STATES.Add(1)
	pe.SET_PIN("WAIT", PIN_HIGH)
}
func (pe *PIN_EMULATOR) CLEAR_WAIT_STATES() {
	pe.WAIT_STATES.Store(0)
	pe.SET_PIN("WAIT", PIN_LOW)
}

// bus abriratonrateration
func (pe *PIN_EMULATOR) REQUEST_BUS_ACCESS(device_id string) bool {
	current_owner := pe.BUS_OWNER.Load().(string)
	if current_owner != "CPU" {
		return false // someone else already stole the bus
	}
	// signal hold request
	pe.SET_PIN("HOLD", PIN_HIGH)
	if pe.GET_PIN("HLDA") == PIN_HIGH {
		pe.BUS_OWNER.Store(device_id)
		pe.SET_BUS_TO_HIGH_Z()
		return true
	}
	return false
}
func (pe *PIN_EMULATOR) ACKNOWLEDGE_BUS_REQUEST() {
	pe.SET_PIN("HLDA", PIN_HIGH)
	pe.SET_BUS_TO_HIGH_Z()
}
func (pe *PIN_EMULATOR) RELEASE_BUS(device_id string) {
	current_owner := pe.BUS_OWNER.Load().(string)
	if current_owner == device_id {
		pe.SET_PIN("HOLD", PIN_LOW)
		pe.SET_PIN("HLDA", PIN_LOW)
		pe.BUS_OWNER.Store("CPU")
		// cpu takes back control of the bus
	}
}
func (pe *PIN_EMULATOR) SET_BUS_TO_HIGH_Z() {
	// set all bus lines to high (nobody drives them)
	for i := 0; i < 16; i++ {
		pe.SET_PIN(MAP_ADDRESS_PIN_NAME(i), PIN_Z)
	}
	for i := 0; i < 8; i++ {
		pe.SET_PIN(MAP_DATA_PIN_NAME(i), PIN_Z)
	}
}
func (pe *PIN_EMULATOR) HANDLE_HALT_STATE() {
	// this would stop the clock or something in real hardware
}

// clock generation
func (pe *PIN_EMULATOR) GENERATE_CLOCK() {
	lastFreq := pe.CLOCK_FREQ_HZ.Load()
	period := time.Second / time.Duration(lastFreq)
	half := period / 2

	ticker := time.NewTicker(half)
	defer ticker.Stop()
	pe.CLOCK_RUNNING.Store(true)
	phase := false

	for {
		select {
		case <-pe.CLOCK_CTX.Done():
			pe.CLOCK_RUNNING.Store(false)
			return
		case <-ticker.C:
			freq := pe.CLOCK_FREQ_HZ.Load()
			if freq != lastFreq {
				lastFreq = freq
				ticker.Stop()
				period = time.Second / time.Duration(freq)
				half = period / 2
				ticker = time.NewTicker(half)
				continue
			}

			if phase {
				pe.SET_PIN("PHI1", PIN_LOW)
				pe.SET_PIN("PHI2", PIN_HIGH)
				select {
				case pe.CLOCK_PHI2 <- true:
				default:
				}
			} else {
				pe.SET_PIN("PHI1", PIN_HIGH)
				pe.SET_PIN("PHI2", PIN_LOW)
				select {
				case pe.CLOCK_PHI1 <- true:
				default:
				}
			}
			phase = !phase
		}
	}
}
func (pe *PIN_EMULATOR) SET_CLOCK_FREQUENCY(freq_hz uint64) {
	pe.CLOCK_FREQ_HZ.Store(freq_hz)
	// would need to restart clock generator or something in real hardware but i lazy to implement
}
func (pe *PIN_EMULATOR) STOP_CLOCK() {
	if pe.CLOCK_CANCEL != nil {
		pe.CLOCK_CANCEL()
	}
	pe.CLOCK_RUNNING.Store(false)
}

// i gate
func (pe *PIN_EMULATOR) SIMULATE_PROPAGATION_DELAY() {
	if pe.TIMING_ENABLED.Load() {
		time.Sleep(time.Duration(PROPAGATION_DELAY_NS) * time.Nanosecond)
		pe.LAST_TRANSITION.Store(time.Now().UnixNano())
	}
}
func (pe *PIN_EMULATOR) SIMULATE_SETUP_TIME() {
	if pe.TIMING_ENABLED.Load() {
		time.Sleep(time.Duration(SETUP_TIME_NS) * time.Nanosecond)
	}
}
func (pe *PIN_EMULATOR) SIMULATE_HOLD_TIME() {
	if pe.TIMING_ENABLED.Load() {
		time.Sleep(time.Duration(HOLD_TIME_NS) * time.Nanosecond)
	}
}
func (pe *PIN_EMULATOR) SIMULATE_ACCESS_TIME() {
	if pe.TIMING_ENABLED.Load() {
		// simulate memory/io access time
		time.Sleep(time.Duration(100) * time.Nanosecond)
	}
}
func (pe *PIN_EMULATOR) ENABLE_TIMING_SIMULATION(enable bool) {
	pe.TIMING_ENABLED.Store(enable)
}

// reset handling
func (pe *PIN_EMULATOR) RESET_SYSTEM() {
	pe.DEVICE_MUTEX.RLock()
	for _, device := range pe.CONNECTED_DEVICES {
		device.RESET() // let each device handle its own reset logic
	}
	pe.DEVICE_MUTEX.RUnlock()
	// reset cpu state to power-on defaults
	pe.SET_ADDRESS_BUS(0x0000)
	pe.SET_DATA_BUS(0x00)
	pe.BUS_OWNER.Store("CPU")
	pe.PROCESSOR_STATE.Store(uint8(0))
	pe.MACHINE_CYCLE.Store(0)
	pe.INSTRUCTION_CYCLE.Store(0)
	pe.PENDING_INTERRUPTS.Store(0)
	pe.WAIT_STATES.Store(0)
	// clear all control signals except reset
	pe.SET_PIN("SYNC", PIN_LOW)
	pe.SET_PIN("DBIN", PIN_LOW)
	pe.SET_PIN("WR", PIN_HIGH)    // write is active low
	pe.SET_PIN("READY", PIN_HIGH) // ready is active high
	pe.SET_PIN("WAIT", PIN_LOW)
	pe.SET_PIN("HOLD", PIN_LOW)
	pe.SET_PIN("HLDA", PIN_LOW)
	pe.SET_PIN("INT", PIN_LOW)
	pe.SET_PIN("INTE", PIN_LOW) // interrupts disabled after reset
}

/*
	do
	NOT
	THE
	FUCK
	UDHGAIUSDDBGASI
*/
func (pe *PIN_EMULATOR) GET_MACHINE_CYCLE_COUNT() uint32 {
	return pe.MACHINE_CYCLE.Load()
}
func (pe *PIN_EMULATOR) GET_INSTRUCTION_CYCLE_COUNT() uint32 {
	return pe.INSTRUCTION_CYCLE.Load()
}
func (pe *PIN_EMULATOR) GET_PROCESSOR_STATE() uint8 {
	return pe.PROCESSOR_STATE.Load().(uint8)
}
func (pe *PIN_EMULATOR) GET_BUS_OWNER() string {
	return pe.BUS_OWNER.Load().(string)
}
func (pe *PIN_EMULATOR) IS_CLOCK_RUNNING() bool {
	return pe.CLOCK_RUNNING.Load()
}
func (pe *PIN_EMULATOR) GET_WAIT_STATE_COUNT() uint32 {
	return pe.WAIT_STATES.Load()
}

/* useful if debug ig
func (pe *PIN_EMULATOR) DUMP_PIN_STATE() map[string]PIN_STATE {
	state := make(map[string]PIN_STATE)
	for pin_name := range PIN_MAP {
		state[pin_name] = pe.GET_PIN(pin_name)
	}
	return state
}
*/
// btw after 3 hour of making this shit i realize uhhhh i dunno can fix later
