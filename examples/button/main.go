package main

import (
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"github.com/warmans/go-rtk"
	"log"
	"time"
)

// this is just a microswitch with one side connected to 3.3v and the other connected to pin 10.

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

	if err := client.Setup(pin, rtk.PinModeInput, rtk.Pull(rtk.PullDown)); err != nil {
		log.Fatalf("setup failed %v", err)
	}

	fmt.Println("Waiting...")

	pinState := rtk.PinStateLow
	for {
		val, err := client.Input(pin)
		if err != nil {
			log.Fatalf("output failed %v", err)
		}
		if val != pinState {
			pinState = val
			fmt.Printf("Pin changed to: %v", val)
		}
		time.Sleep(time.Millisecond)
	}
}
