package max6675

import (
	"errors"
	"time"

	"github.com/gregorychen3/espresso-controller/internal/espresso/temperature"
	"github.com/stianeikeland/go-rpio/v4"
)

const dataBitLength = 16

type Max6675 struct {
	cs   rpio.Pin
	clk  rpio.Pin
	miso rpio.Pin
}

func NewMax6675(csPin, clkPin, misoPin int) *Max6675 {
	c := Max6675{
		cs:   rpio.Pin(csPin),
		clk:  rpio.Pin(clkPin),
		miso: rpio.Pin(misoPin),
	}
	c.cs.Output()
	c.clk.Output()
	c.miso.Input()
	return &c
}

func (m *Max6675) Sample() (*temperature.Sample, error) {
	bits := m.readBits()
	if err := checkErr(bits); err != nil {
		return nil, err
	}
	return &temperature.Sample{
		Value:      bitsToTemperature(bits),
		ObservedAt: time.Now(),
	}, nil
}

func (m *Max6675) readBits() uint32 {
	m.cs.Low()        // begin the read
	defer m.cs.High() // end the read

	var bits uint32
	for i := 0; i < dataBitLength; i++ {
		m.clk.High()
		bit := m.miso.Read()
		if bit == rpio.High {
			bits |= 0x1
		}
		if i != dataBitLength-1 { // shift left to get to the next bit to be read
			bits <<= 1
		}
		m.clk.Low() // pulse low, then high again to get the next bit
	}
	return bits
}

func bitsToTemperature(bits uint32) float32 {
	//log.Info(fmt.Sprintf("temperature bits %016b", bits))
	thermoData := bits >> 3 // only use 12 bits
	result := float32(thermoData) * 0.25 // 12 bits = 4096 possible temperature, range of 0 - 1024 degrees, 0.25 degree per unit
	return result
}

func checkErr(bits uint32) error {
	openCircuit := bits&0b100 == 1 // fault bit D2
	if openCircuit {
		return errors.New("open circuit")
	} 
	return nil
}
