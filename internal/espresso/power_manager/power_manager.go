package power_manager

import (
	"github.com/stianeikeland/go-rpio/v4"
	"time"
)

type PowerOnInterval struct {
	From int
	To   int
}

type PowerSchedule struct {
	Frames map[time.Weekday][]PowerOnInterval
}

type PowerManager struct {
	PowerSchedule       PowerSchedule
	AutoOffDuration     time.Duration
	ScheduleOn          bool
	OnSince             time.Time
	powerButtonRelayPin rpio.Pin
	powerButtonPin      rpio.Pin
	powerLedPin         rpio.Pin
	machineOn           bool
}

func NewPowerManager(powerSchedule PowerSchedule, autoOffDuration time.Duration, powerButtonRelayPinNum int, powerButtonPinNum int, powerLedPinNum int) *PowerManager {

	powerButtonRelayPin := rpio.Pin(powerButtonRelayPinNum)
	
	powerButtonRelayPin.Output()
	
	powerButtonPin := rpio.Pin(powerButtonPinNum)
	
	powerButtonPin.Input()
	powerButtonPin.PullDown()
	
	powerLedPin := rpio.Pin(powerLedPinNum)
	powerLedPin.Output()

	return &PowerManager{
		PowerSchedule:       powerSchedule,
		AutoOffDuration:     autoOffDuration,
		ScheduleOn:          false,
		OnSince:             time.Time{},
		powerButtonRelayPin: powerButtonRelayPin,
		powerButtonPin:      powerButtonPin,
		powerLedPin:         powerLedPin,
		machineOn:           false,
	}
}

func (p *PowerManager) Run() {
	go func() {
		for {

			currentTime := time.Now()

			if p.inSchedule(currentTime) {
				p.PowerOn()
				p.ScheduleOn = true
			} else {
				if p.ScheduleOn && p.IsMachinePowerOn() {
					p.PowerOff()
					p.ScheduleOn = false
				}
			}

			if p.OnSince.Sub(time.Now()) >= p.AutoOffDuration && !p.ScheduleOn {
				p.PowerOff()
			}

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

func (p *PowerManager) inSchedule(currentTime time.Time) bool {

	hour := currentTime.Hour()
	v, present := p.PowerSchedule.Frames[currentTime.Weekday()]

	if present {
		for _, powerOnInterval := range v {
			if hour >= powerOnInterval.From && hour <= powerOnInterval.To {
				return true
			}
		}
	}

	return false
}

func (p *PowerManager) SetSchedule(newPowerSchedule PowerSchedule) {
	p.PowerSchedule = newPowerSchedule
}

func (p *PowerManager) PowerOn() {
	if p.IsMachinePowerOff() {
		p.powerButtonRelayPin.High()
		p.powerLedPin.High()
		p.machineOn = true
	}
	p.OnSince = time.Now()
}

func (p *PowerManager) PowerOff() {
	if p.IsMachinePowerOn() {
		p.powerButtonRelayPin.Low()
		p.powerLedPin.Low()
		p.machineOn = false
	}
	p.OnSince = time.Time{}
}

func (p *PowerManager) PowerToggle() {
	if p.IsMachinePowerOn() {
		p.PowerOff()
	} else {
		p.PowerOn()
	}
}

func (p *PowerManager) IsMachinePowerOn() bool {
	return p.powerButtonRelayPin.Read() == rpio.High && p.machineOn == true
}

func (p *PowerManager) IsMachinePowerOff() bool {
	return !p.IsMachinePowerOn()
}

func (p *PowerManager) isPowerButtonOn() bool {
	return p.powerButtonPin.Read() == rpio.High
}

func (p *PowerManager) isPowerButtonOff() bool {
	return !p.isPowerButtonOff()
}

func (p *PowerManager) Shutdown() {
	p.powerButtonRelayPin.Low()
}
