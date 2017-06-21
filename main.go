package main

import (
	"fmt"
	"github.com/stianeikeland/go-rpio"
	"os"
	"time"
)

var (
	waterStations = []int{24,25}
)

func main() {

	// Open and map memory to access gpio, check for errors
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// sprinkler
	turnOn(waterStations, 5, 1)

	// garage
	toggle(24, 5)

}

func turnOn(pins []int, min time.Duration, sleep time.Duration) {
	for _, v := range pins {
		pin := rpio.Pin(v)
		pin.Output()
		pin.Low()
		time.Sleep(time.Minute*min)
		pin.High()
		time.Sleep(time.Minute*sleep)
	}
}

func toggle(pinNo int, sec time.Duration) {
	pin := rpio.Pin(pinNo)
	pin.Output()
	pin.Low()
	time.Sleep(time.Second*sec)
	pin.High()
}

