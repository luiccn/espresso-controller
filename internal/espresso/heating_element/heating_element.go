package heating_element

import (
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

type HeatingElement struct {
	heatingElementRelayPin rpio.Pin
	dutyFactor             float32
}

func NewHeatingElement(heatingElementRelayPinNum int) *HeatingElement {
	heatingElementRelayPin := rpio.Pin(heatingElementRelayPinNum)
	heatingElementRelayPin.Output()

	return &HeatingElement{
		heatingElementRelayPin: heatingElementRelayPin,
		dutyFactor:             0,
	}
}

func (h *HeatingElement) Run() {
	go func() {
		for {
			if h.dutyFactor == 0 {
				h.off()
				time.Sleep(1 * time.Second)
				continue
			}

			onMs := h.dutyFactor * 1000
			offMs := (1 - h.dutyFactor) * 1000

			h.on()
			time.Sleep(time.Duration(onMs) * time.Millisecond)
			h.off()
			time.Sleep(time.Duration(offMs) * time.Millisecond)
		}
	}()
}

func (h *HeatingElement) SetDutyFactor(factor float32) {
	h.dutyFactor = factor
}

func (h *HeatingElement) on() {
	h.heatingElementRelayPin.High()
}

func (h *HeatingElement) off() {
	h.heatingElementRelayPin.Low()
}

func (h *HeatingElement) Shutdown() {
	h.heatingElementRelayPin.Low()
}
