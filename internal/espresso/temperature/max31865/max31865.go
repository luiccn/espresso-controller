package max31865

import (
	"fmt"
	"math"
	"time"

	"github.com/luiccn/espresso-controller/internal/espresso/temperature"
	"github.com/luiccn/espresso-controller/internal/log"
	"github.com/stianeikeland/go-rpio/v4"
)

const (
	_CONFIG_REG    uint8 = 0x00
	_RTDMSB_REG    uint8 = 0x01
	_RTDLSB_REG    uint8 = 0x02
	_HFAULTMSB_REG uint8 = 0x03
	_HFAULTLSB_REG uint8 = 0x04
	_LFAULTMSB_REG uint8 = 0x05
	_LFAULTLSB_REG uint8 = 0x06
	_FAULTSTAT_REG uint8 = 0x07
)

const (
	_FAULT_HIGHTHRESH uint8 = 0x80
	_FAULT_LOWTHRESH  uint8 = 0x40
	_FAULT_REFINLOW   uint8 = 0x20
	_FAULT_REFINHIGH  uint8 = 0x10
	_FAULT_RTDINLOW   uint8 = 0x08
	_FAULT_OVUV       uint8 = 0x04
)

const (
	_CONFIG_BIAS      uint8 = 0x80
	_CONFIG_MODEAUTO  uint8 = 0x40
	_CONFIG_MODEOFF   uint8 = 0x00
	_CONFIG_1SHOT     uint8 = 0x20
	_CONFIG_3WIRE     uint8 = 0x10
	_CONFIG_24WIRE    uint8 = 0x00
	_CONFIG_FAULTSTAT uint8 = 0x02
	_CONFIG_FILT50HZ  uint8 = 0x01
	_CONFIG_FILT60HZ  uint8 = 0x00
)

const (
	_RTD_A float32 = 3.9083e-3
	_RTD_B float32 = -5.775e-7
)

const (
	WIRE_2 = 0
	WIRE_3 = 1
	WIRE_4 = 0
)

type Max31865 struct {
	csPin   rpio.Pin
	misoPin rpio.Pin
	mosiPin rpio.Pin
	clkPin  rpio.Pin
}

func (m *Max31865) Sample() (*temperature.Sample, error) {
	t := m.ReadTemperature(100, 430)
	return &temperature.Sample{
		Value:      t,
		ObservedAt: time.Now(),
	}, nil
}

func NewMax31865(cs int, clk int, miso int, mosi int) *Max31865 {

	s := &Max31865{}

	s.csPin = rpio.Pin(cs)
	s.misoPin = rpio.Pin(miso)
	s.mosiPin = rpio.Pin(mosi)
	s.clkPin = rpio.Pin(clk)

	s.csPin.Output()
	s.csPin.High()

	s.clkPin.Output()
	s.clkPin.Low()

	s.misoPin.Input()
	s.mosiPin.Output()

	s.setWires(WIRE_3)
	s.enableBias(false)
	s.autoConvert(false)
	s.clearFault()

	return s
}

func (s *Max31865) ReadTemperature(RTDnominal float32, refResistor float32) float32 {

	Rt := float32(s.ReadRTD())

	log.Info(fmt.Sprintf("Rt %f Fault %d", Rt, s.readFault()))


	Rt /= 32768
	Rt *= refResistor

	log.Info(fmt.Sprintf("Rt %f Fault %d", Rt, s.readFault()))

	Z1 := -_RTD_A
	Z2 := _RTD_A*_RTD_A - (4 * _RTD_B)
	Z3 := (4 * _RTD_B) / RTDnominal
	Z4 := 2 * _RTD_B

	temp := Z2 + (Z3 * Rt)
	temp = (float32(math.Sqrt(float64(temp))) + Z1) / Z4

	if temp >= 0 {
		return temp
	}

	Rt /= RTDnominal
	Rt *= 100

	rpoly := Rt

	temp = -242.02
	temp += 2.2228 * rpoly
	rpoly *= Rt
	temp += 2.5859e-3 * rpoly
	rpoly *= RTDnominal
	temp -= 4.8260e-6 * rpoly
	rpoly *= Rt
	temp -= 2.8183e-8 * rpoly
	rpoly *= Rt
	temp += 1.5243e-10 * rpoly

	return temp
}

func (s *Max31865) readFault() uint8 {
	return s.read8(_FAULTSTAT_REG)
}

func (s *Max31865) clearFault() {
	t := s.read8(_CONFIG_REG)
	t &= ^uint8(0x2C)
	t |= _CONFIG_FAULTSTAT
	s.write8(_CONFIG_REG, t)
}

func (s *Max31865) enableBias(b bool) {
	t := s.read8(_CONFIG_REG)
	if b {
		t |= _CONFIG_BIAS
	} else {
		t &= ^_CONFIG_BIAS
	}
	s.write8(_CONFIG_REG, t)
}

func (s *Max31865) autoConvert(b bool) {
	t := s.read8(_CONFIG_REG)
	if b {
		t |= _CONFIG_MODEAUTO
	} else {
		t &= ^_CONFIG_MODEAUTO
	}
	s.write8(_CONFIG_REG, t)
}

func (s *Max31865) setWires(wires int) {
	t := s.read8(_CONFIG_REG)
	if wires == WIRE_3 {
		t |= _CONFIG_3WIRE
	} else {
		t &= ^_CONFIG_3WIRE
	}
	s.write8(_CONFIG_REG, t)
}

func (s *Max31865) ReadRTD() uint16 {

	s.clearFault()
	s.enableBias(true)
	time.Sleep(10 * time.Millisecond)

	t := s.read8(_CONFIG_REG)
	t |= _CONFIG_1SHOT
	s.write8(_CONFIG_REG, t)
	time.Sleep(65 * time.Millisecond)

	rtd := s.read16(_RTDMSB_REG)
	rtd >>= 1
	return rtd

}

func (s *Max31865) write(addr uint8, v []uint8) {
	addr |= 0x80
	s.clkPin.Low()
	s.csPin.Low()

	s.transfer8(addr)
	for n := 0; n < len(v); n++ {
		s.transfer8(v[n])
	}
	s.csPin.High()
}

func (s *Max31865) write8(addr uint8, v uint8) {
	s.write(addr, []uint8{v})
}

func (s *Max31865) read(addr uint8, v []uint8) {
	addr &= 0x7F
	s.clkPin.Low()
	s.csPin.Low()
	s.transfer8(addr)
	for n := 0; n < len(v); n++ {
		v[n] = s.transfer8(0xFF)
	}
	s.csPin.High()
}

func (s *Max31865) read8(addr uint8) uint8 {
	var v = make([]uint8, 1)
	s.read(addr, v)
	return v[0]
}

func (s *Max31865) read16(addr uint8) uint16 {
	var v = make([]uint8, 2)
	s.read(addr, v)
	return (uint16(v[0]) << 8) | uint16(v[1])
}

func (s *Max31865) transfer8(v uint8) uint8 {
	var reply uint8 = 0
	for i := 7; i >= 0; i-- {
		reply <<= 1
		s.clkPin.High()
		bv := v & (1 << uint(i))

		if bv != 0 {
			s.mosiPin.High()
		} else {
			s.mosiPin.Low()
		}

		s.clkPin.Low()

		if s.misoPin.Read() == rpio.High {
			reply |= 1
		}
	}
	return reply
}
