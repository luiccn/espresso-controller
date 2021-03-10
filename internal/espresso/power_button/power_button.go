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
			if p.isPowerButtonOn() {
				p.PowerToggle()
				time.Sleep(1000 * time.Millisecond)
				for p.isPowerButtonOn() {
					time.Sleep(1000 * time.Millisecond)
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func (p *PowerButton) PowerOn() {
	if p.IsMachinePowerOff() {
		p.powerButtonRelayPin.High()
	p.machineOn = true	
	}
}

func (p *PowerButton) PowerOff() {
	if p.IsMachinePowerOn() {
		p.powerButtonRelayPin.Low()
	p.machineOn = false	
	}
}

func (p *PowerButton) PowerToggle() {
	if p.IsMachinePowerOn() {
		p.PowerOff()
	} else {
		p.PowerOn()	
	}
}

func (p *PowerButton) isPowerButtonOn() bool {
	return p.powerButtonPin.Read() == rpio.High
}

func (p *PowerButton) isPowerButtonOff() bool {
	return !p.isPowerButtonOff()
}

func (p *PowerButton) IsMachinePowerOn() bool {
	return p.powerButtonRelayPin.Read() == rpio.High && p.machineOn == true
}

func (p *PowerButton) IsMachinePowerOff() bool {
	return !p.IsMachinePowerOn()
}

func (p *PowerButton) Shutdown() {
	p.powerButtonRelayPin.Low()
}
