package main

import (
	"github.com/jacobsa/go-serial/serial"
	"github.com/warmans/go-rtk"
	"log"
	"time"
)

const pin = uint8(10)

func main() {

	// Open the port.
	port, err := serial.Open(rtk.SerialOptions("/dev/serial/by-path/pci-0000:00:06.0-usb-0:2:1.0-port0"))
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	defer port.Close()

	client := rtk.NewGPIOClient(port)
	defer client.Close()

	if err := client.Setup(pin, rtk.InitialPinMode(rtk.PinModeOutput), rtk.Pull(rtk.PullDown), rtk.InitialState(rtk.PinStateHigh)); err != nil {
		log.Fatalf("setup failed %v", err)
	}

	state := rtk.PinStateLow
	for i := 0; i < 6; i++ {
		if i%2 == 0 {
			state = rtk.PinStateLow
		} else {
			state = rtk.PinStateHigh
		}
		log.Printf("Turning on: %v\n", state)
		if err := client.Output(pin, state); err != nil {
			log.Fatalf("output failed %v", err)
		}
		time.Sleep(time.Second)
	}
}
