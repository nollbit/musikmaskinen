package controller

import (
	"errors"
	"io"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/jacobsa/go-serial/serial"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	controllerPortFlag = kingpin.Flag("controller-port", "Serial port for the MM controller").String()
)

const (
	// CommandLedOff : Turn off the controller led
	CommandLedOff = byte('O')

	// CommandLedBlink : Blink the controller led
	CommandLedBlink = byte('B')

	// CommandLedGlow : glow the controller led
	CommandLedGlow = byte('G')

	// EventCmdRotaryEncoderClockwise means that the rotary knob was turned clockwise one stop
	EventCmdRotaryEncoderClockwise = byte('W')

	// EventCmdRotaryEncoderCounterClockwise means that the rotary knob was turned counter-clockwise one stop
	EventCmdRotaryEncoderCounterClockwise = byte('C')

	// EventCmdRotaryEncoderButton means that the rotary knob was wash pushed
	EventCmdRotaryEncoderButton = byte('D')

	// EventCmdPushButton means that the push button was pushed
	EventCmdPushButton = byte('P')
)

type (
	Controller struct {
		port          io.ReadWriteCloser
		CommandEvents chan byte
		Errs          chan error
	}
)

func (c *Controller) WriteCommand(b byte) error {
	if c.port == nil {
		log.Debugf("Dummy controller not sending command %b", b)
		return nil
	}
	log.Debugf("Sending command %b to controller", b)
	_, err := c.port.Write([]byte{b})
	return err
}

func (c *Controller) Close() {
	if c.port != nil {
		c.port.Close()
	}
}

func NewController() (*Controller, error) {

	log.Infof("controllerPortFlag = %s", *controllerPortFlag)

	var controllerPort string
	if *controllerPortFlag == "" {
		cp, err := findPortCandidate()
		if err != nil {
			return nil, err
		}
		controllerPort = cp
	} else {
		controllerPort = *controllerPortFlag
	}

	log.Infof("Opening controller port %s", controllerPort)

	options := serial.OpenOptions{
		PortName:        controllerPort,
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 1,
	}

	port, err := serial.Open(options)
	if err != nil {
		return nil, err
	}

	commandChan := make(chan byte)
	errChan := make(chan error)
	controller := &Controller{
		port:          port,
		CommandEvents: commandChan,
		Errs:          errChan,
	}

	go func() {
		buf := make([]byte, 8)
		for {

			n, err := port.Read(buf)
			if err != nil {
				log.WithError(err).Error("Error reading from controller port")
				errChan <- err
			}

			for i := 0; i < n; i++ {
				b := buf[i]

				log.Debugf("Got command %b from controller", b)
				commandChan <- b
			}
		}
	}()

	return controller, nil
}

// NewDummyController creates a controller that never emits events
func NewDummyController() *Controller {
	commandChan := make(chan byte)
	errChan := make(chan error)
	controller := &Controller{
		port:          nil,
		CommandEvents: commandChan,
		Errs:          errChan,
	}

	return controller
}

func findPortCandidate() (string, error) {
	// TODO(nollbit) make this work on non-darwin
	matches, err := filepath.Glob("/dev/cu.usbmodem*")
	if err != nil {
		return "", err
	}

	if len(matches) > 0 {
		return matches[0], nil
	}

	return "", errors.New("Can't find port candidate for controller")

}
