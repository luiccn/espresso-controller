package power_manager

import (
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

type PowerOnInterval struct {
	From int
	To   int
}

type PowerSchedule struct {
	Frames map[time.Weekday][]PowerOnInterval
}

type PowerManager struct {
	PowerSchedule        PowerSchedule
	AutoOffDuration      time.Duration
	powerRelayPin        rpio.Pin
	powerButtonPin       rpio.Pin
	OnSince              time.Time
	CurrentlyInASchedule bool
	LastInteraction      string
	StopScheduling       bool
	currentSchedule      PowerOnInterval
	totalOff             bool
}

type PowerManagerStatus struct {
	PowerSchedule        PowerSchedule
	AutoOffDuration      time.Duration
	OnSince              time.Time
	CurrentlyInASchedule bool
	LastInteraction      string
	PowerOn              bool
	StopScheduling       bool
	TotalOff             bool
}

func NewPowerManager(powerSchedule PowerSchedule, autoOffDuration time.Duration, powerRelayPinNum int, powerButtonPinNum int, powerLedPinNum int) *PowerManager {

	powerRelayPin := rpio.Pin(powerRelayPinNum)
	powerRelayPin.Output()
	powerRelayPin.Low()

	powerButtonPin := rpio.Pin(powerButtonPinNum)
	powerButtonPin.Input()
	powerButtonPin.PullDown()

	powerLedPin := rpio.Pin(powerLedPinNum)
	powerLedPin.Output()
	powerLedPin.Low()

	return &PowerManager{
		PowerSchedule:        powerSchedule,
		AutoOffDuration:      autoOffDuration,
		CurrentlyInASchedule: false,
		OnSince:              time.Time{},
		powerRelayPin:        powerRelayPin,
		powerButtonPin:       powerButtonPin,
		LastInteraction:      "Start-up Off",
		StopScheduling:       false,
		currentSchedule:      PowerOnInterval{},
		totalOff:             false,
	}
}

func (p *PowerManager) Run() {
	go func() {
		for {

			currentTime := time.Now()

			p.powerButtonBehaviour()

			if p.totalOff {
				continue
			}

			p.powerScheduleBehaviour(currentTime)
			p.autoOffBehaviour()

			time.Sleep(200 * time.Millisecond)
		}
	}()
}

func (p *PowerManager) autoOffBehaviour() {
	if !p.OnSince.Equal(time.Time{}) && time.Now().Sub(p.OnSince) >= p.AutoOffDuration && !p.CurrentlyInASchedule {
		p.powerOff()
		p.LastInteraction = "Auto-off"
	}
}

func (p *PowerManager) powerScheduleBehaviour(currentTime time.Time) {
	powerOnInterval, inSchedule := p.inSchedule(currentTime)
	if inSchedule && p.currentSchedule != powerOnInterval {
		p.StopScheduling = false
		p.currentSchedule = powerOnInterval
	}
	if !p.StopScheduling {
		if inSchedule {
			p.powerOn()
			p.CurrentlyInASchedule = true
			p.LastInteraction = "Scheduled Power On"
		} else {
			if p.CurrentlyInASchedule && p.IsMachinePowerOn() {
				p.powerOff()
				p.CurrentlyInASchedule = false
				p.LastInteraction = "Scheduled Power Off"
			}
		}
	} else {
		p.CurrentlyInASchedule = false
	}
}

func (p *PowerManager) powerButtonBehaviour() {
	if p.isPowerButtonOn() {
		count := 0
		for p.isPowerButtonOn() && count < 10 {
			count++
			time.Sleep(100 * time.Millisecond)
		}

		if p.IsMachinePowerOn() {
			p.powerOff()
			p.LastInteraction = "Power Button Off"
			if p.CurrentlyInASchedule {
				p.StopScheduling = true
			}
		} else {
			p.powerOn()
			p.totalOff = false
			p.LastInteraction = "Power Button On"
		}

		for p.isPowerButtonOn() {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (p *PowerManager) GetStatus() PowerManagerStatus {
	return PowerManagerStatus{
		PowerSchedule:        p.PowerSchedule,
		AutoOffDuration:      p.AutoOffDuration,
		OnSince:              p.OnSince,
		CurrentlyInASchedule: p.CurrentlyInASchedule,
		LastInteraction:      p.LastInteraction,
		PowerOn:              p.IsMachinePowerOn(),
		StopScheduling:       p.StopScheduling,
		TotalOff:             p.totalOff,
	}
}

func (p *PowerManager) SetSchedule(newPowerSchedule PowerSchedule) {
	p.PowerSchedule = newPowerSchedule
}

func (p *PowerManager) powerOn() {
	if p.IsMachinePowerOff() {
		p.powerRelayPin.High()
		p.OnSince = time.Now()
		p.LastInteraction = "Power On Call"
	}
}

func (p *PowerManager) powerOff() {
	if p.IsMachinePowerOn() {
		p.powerRelayPin.Low()
		p.OnSince = time.Time{}
		p.LastInteraction = "Power Off Call"
	}
}

func (p *PowerManager) PowerOn() {
	p.totalOff = false
	p.powerOn()
}

func (p *PowerManager) PowerOff() {
	if p.CurrentlyInASchedule {
		p.StopScheduling = true
	}
	p.powerOff()
}

func (p *PowerManager) TotalPowerOff() {
	if p.CurrentlyInASchedule {
		p.StopScheduling = true
	}
	p.powerOff()
	p.LastInteraction = "Total Power Off Call"
	p.totalOff = true
}

func (p *PowerManager) ScheduleOn() {
	p.StopScheduling = false
}

func (p *PowerManager) ScheduleOff() {
	p.StopScheduling = true
}

func (p *PowerManager) PowerToggle() {
	if p.IsMachinePowerOn() {
		p.PowerOff()
	} else {
		p.PowerOn()
	}
}

func (p *PowerManager) IsMachinePowerOn() bool {
	return p.powerRelayPin.Read() == rpio.High
}

func (p *PowerManager) IsMachinePowerOff() bool {
	return !p.IsMachinePowerOn()
}

func (p *PowerManager) isPowerButtonOn() bool {
	return p.powerButtonPin.Read() == rpio.High
}

func (p *PowerManager) inSchedule(currentTime time.Time) (PowerOnInterval, bool) {

	hour := currentTime.Hour()
	v, present := p.PowerSchedule.Frames[currentTime.Weekday()]

	if present {
		for _, powerOnInterval := range v {
			if hour >= powerOnInterval.From && hour <= powerOnInterval.To {
				return powerOnInterval, true
			}
		}
	}

	return p.currentSchedule, false
}

func (p *PowerManager) Shutdown() {
	p.powerRelayPin.Low()
}
