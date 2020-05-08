package rtk

import (
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"io"
	"time"
)

var pinMap = map[uint8]uint8{
	3:  2,
	5:  3,
	7:  4,
	8:  14,
	10: 15,
	11: 17,
	12: 18,
	13: 27,
	15: 22,
	16: 23,
	18: 24,
	19: 10,
	21: 9,
	22: 25,
	23: 11,
	24: 8,
	26: 7,
	27: 0,
	28: 1,
	29: 5,
	31: 6,
	32: 12,
	33: 13,
	35: 19,
	36: 16,
	37: 26,
	38: 20,
	40: 21,
}

type PullMode string

func (m PullMode) Ptr() *PullMode {
	return &m
}

const (
	PullUp   PullMode = "U"
	PullDown PullMode = "D"
	PullNone PullMode = "N"
)

type PinMode string

func (m PinMode) Ptr() *PinMode {
	return &m
}

const (
	PinModeInput  PinMode = "I"
	PinModeOutput PinMode = "O"
)

type PinState string

func (s PinState) Ptr() *PinState {
	return &s
}

const (
	PinStateHigh PinState = "1"
	PinStateLow  PinState = "0"
)

func NewGPIOClient(port io.ReadWriteCloser) *GPIOClient {
	return &GPIOClient{port: port}
}

// serial client for GPIO device
type GPIOClient struct {
	port      io.ReadWriteCloser
	boardMode uint8
}

func (g *GPIOClient) SetMode(mode uint8) error {
	if mode != 0 && mode != 1 {
		return fmt.Errorf("%d is not a valid board mode", mode)
	}
	g.boardMode = mode
	return nil
}

func (g *GPIOClient) Output(pin uint8, state PinState) error {
	if err := g.validatePin(pin); err != nil {
		return err
	}
	if err := g.write(g.pinch(pin)); err != nil {
		return err
	}
	if err := g.write(string(state)); err != nil {
		return err
	}
	return nil
}

func (g *GPIOClient) Input(pin uint8) (PinState, error) {
	if err := g.validatePin(pin); err != nil {
		return "", err
	}
	if err := g.write(g.pinch(pin) + "?"); err != nil {
		return "", err
	}
	payload := []byte{}

	// max payload is 4 bytes
	for len(payload) < 4 {
		buff := make([]byte, 1)
		read, err := g.port.Read(buff)
		if err != nil {
			return "", err
		}
		if read == 1 {
			// this doesn't seem to make sense. In the python library it will retry if there
			// are less than 4 bytes. I'm not sure why the newline would terminate a message at fewer
			// than 4 bytes just to discard it later.
			if string(buff[0]) == "\n" {
				break
			}
			payload = append(payload, buff...)
		} else {
			time.Sleep(time.Millisecond)
		}
	}

	// should have a max 4 byte payload.
	// I'm going to assume if we have at least 2 bytes then it's still
	// a valid message because it only needs the second byte to get the state
	// of the pin.

	if len(payload) > 1 {
		switch string(payload[1]) {
		case "0":
			return PinStateLow, nil
		case "1":
			return PinStateHigh, nil
		default:
			return "", fmt.Errorf("unknown pin state %d", payload[1])
		}
	} else {
		return "", fmt.Errorf("payload was terminated prematurely")
	}
}

func (g *GPIOClient) Setup(pin uint8, opts ...SetupOpt) error {
	if err := g.validatePin(pin); err != nil {
		return err
	}
	setupOpts := &setupOpts{}
	for _, opt := range opts {
		opt(setupOpts)
	}
	if setupOpts.pinMode != nil {
		if err := g.write(g.pinch(pin) + string(*setupOpts.pinMode)); err != nil {
			return err
		}
	}
	if setupOpts.pull != nil {
		if err := g.write(g.pinch(pin) + string(*setupOpts.pull)); err != nil {
			return err
		}
	}
	if setupOpts.initialState != nil {
		if err := g.Output(pin, *setupOpts.initialState); err != nil {
			return err
		}
	}
	return nil
}

// Not sure what close is supposed to do. In the real client id seems to do nothing.
// So I'll just set all the pins to off.
func (g *GPIOClient) Close() {
	for bm0, bm1 := range pinMap {
		if g.boardMode == 0 {
			g.Setup(bm0, InitialPinMode(PinModeOutput), Pull(PullNone), InitialState(PinStateLow))
		} else {
			g.Setup(bm1, InitialPinMode(PinModeOutput), Pull(PullNone), InitialState(PinStateLow))
		}
	}
}

func (g *GPIOClient) pinch(pin uint8) string {
	if g.boardMode == 0 {
		return string(pin + uint8('a'))
	}
	panic("board mode 1 not implemented")
}

func (g *GPIOClient) write(data string) error {
	_, err := g.port.Write([]byte(data))
	return err
}

func (g *GPIOClient) validatePin(pin uint8) error {
	for bm0, bm1 := range pinMap {
		if g.boardMode == 0 {
			if pin == bm0 {
				return nil
			}
		} else {
			if pin == bm1 {
				return nil

			}
		}
	}
	return fmt.Errorf("%d is not a valid pin for this board mode", pin)
}

type setupOpts struct {
	pull         *PullMode
	initialState *PinState
	pinMode      *PinMode
}

type SetupOpt func(opts *setupOpts)

func Pull(mode PullMode) SetupOpt {
	return func(opts *setupOpts) {
		opts.pull = mode.Ptr()
	}
}

func InitialState(state PinState) SetupOpt {
	return func(opts *setupOpts) {
		opts.initialState = state.Ptr()
	}
}

func InitialPinMode(mode PinMode) SetupOpt {
	return func(opts *setupOpts) {
		opts.pinMode = mode.Ptr()
	}
}

// These are apparently the right settings to talk to the board over a serial connection.
// see examples.
func SerialOptions(port string) serial.OpenOptions {
	return serial.OpenOptions{
		// this will be wrong for you. Look in /dev/serial/by-path for your device.
		PortName:        port,
		BaudRate:        230400,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}
}
