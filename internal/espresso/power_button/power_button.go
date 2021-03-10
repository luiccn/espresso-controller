package power_button

import (
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

type PowerButton struct {
	powerButtonRelayPin rpio.Pin
	powerButtonPin      rpio.Pin
	machineOn           bool
	powerOverrideOn     bool
	powerOverrideOff    bool
}

func NewPowerButton(powerButtonRelayPinNum int, powerButtonPinNum int) *PowerButton {
	powerButtonRelayPin := rpio.Pin(powerButtonRelayPinNum)
	powerButtonPin := rpio.Pin(powerButtonPinNum)

	powerButtonRelayPin.Output()

	powerButtonPin.Input()
	powerButtonPin.PullDown()

	return &PowerButton{
		powerButtonRelayPin: powerButtonRelayPin,
		powerButtonPin:      powerButtonPin,
		machineOn:           false,
	}
}

func (p *PowerButton) Run() {
	go func() {
		for {
			if p.powerButtonPin.Read() == rpio.High && p.machineOn == false {
				p.PowerOn()
				time.Sleep(1000 * time.Millisecond)
			} else if p.powerButtonPin.Read() == rpio.High && p.machineOn == true {
				p.PowerOff()			
				time.Sleep(1000 * time.Millisecond)
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func (p *PowerButton) PowerOn() {
	if p.isOff() {
		p.on()
	}
}

func (p *PowerButton) PowerOff() {
	if p.isOn() {
		p.off()
	}
}

func (p *PowerButton) PowerToggle() {
	if p.isOn() {
		p.off()
	} else {
		p.on()	
	}
}

func (p *PowerButton) on() {
	p.powerButtonRelayPin.High()
	p.machineOn = true	
}

func (p *PowerButton) off() {
	p.powerButtonRelayPin.Low()
	p.machineOn = false	
}

func (p *PowerButton) isOn() bool {
	return p.powerButtonRelayPin.Read() == rpio.High && p.machineOn == true
}

func (p *PowerButton) isOff() bool {
	return !p.isOn()
}

func (p *PowerButton) Shutdown() {
	p.powerButtonRelayPin.Low()
}
