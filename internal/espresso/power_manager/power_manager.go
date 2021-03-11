package power_manager

import (
	"time"

	"github.com/gregorychen3/espresso-controller/internal/espresso/power_button"
)

type PowerOnInterval struct {
	from time.Time
	to time.Time
}

type ScheduleFrame struct {
	day time.Weekday
	powerOn []PowerOnInterval
}

type PowerSchedule struct{
	frames map[time.Weekday]ScheduleFrame
}

type PowerManager struct {
	powerButton     *power_button.PowerButton
	powerSchedule   PowerSchedule
	autoOffDuration time.Duration
	scheduleOn bool
	onSince time.Time
}

func NewPowerManager(powerButton *power_button.PowerButton, powerSchedule PowerSchedule, autoOffDuration time.Duration) *PowerManager {
	return &PowerManager{
		powerButton: powerButton,
		powerSchedule: powerSchedule,
		autoOffDuration: autoOffDuration,
		scheduleOn: false,
		onSince: time.Time{},
	}
}

func (p *PowerManager) Run() {
	go func() {
		for {

			currentTime := time.Now()

			if p.inSchedule(currentTime) {
				p.PowerOn()
				p.scheduleOn = true
			} else {
				if p.scheduleOn && p.IsMachinePowerOn() {
					p.PowerOff()
					p.scheduleOn = false
				}
			}

			if p.onSince.Sub(time.Now()) >= p.autoOffDuration && !p.scheduleOn  {
				p.PowerOff()
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func (p *PowerManager) inSchedule(currentTime time.Time) bool {
	
	v, present := p.powerSchedule.frames[currentTime.Weekday()]

	if present {
		for _,powerOnInterval := range v.powerOn {
			if currentTime.After(powerOnInterval.from) && currentTime.Before(powerOnInterval.to) {
				return true
			}
		}
	} 

	return false
}

func (p *PowerManager) PowerOn() {
	p.powerButton.PowerOn()
	p.onSince = time.Now() 
}

func (p *PowerManager) PowerOff() {
	p.powerButton.PowerOff()
	p.onSince = time.Time{}
}

func (p *PowerManager) PowerToggle() {
	if p.IsMachinePowerOn() {
		p.PowerOff()
	} else {
		p.PowerOn()	
	}
}

func (p *PowerManager) IsMachinePowerOn() bool {
	return p.powerButton.IsMachinePowerOn()
}

func (p *PowerManager) IsMachinePowerOff() bool {
	return p.powerButton.IsMachinePowerOff()
}

func (p *PowerManager) Shutdown() {
	p.powerButton.Shutdown()
}
