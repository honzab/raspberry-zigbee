package main

import (
	"flag"
	"image"
	"image/png"
	"log"
	"os"

	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/devices/v3/inky"
	"periph.io/x/host/v3"
)

// sudo raspi-config nonint do_i2c 0
// sudo raspi-config nonint do_spi 0

// https://github.com/periph/devices/blob/f007d15374363a90b9622a089b17cc56616a4f84/inky/inky.go#L209

// https://github.com/gonum/plot/wiki/Example-plots

func initialiseInkyWhat() *inky.Dev {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	b, err := spireg.Open("SPI0.0")
	if err != nil {
		log.Fatal(err)
	}

	dc := gpioreg.ByName("22")
	reset := gpioreg.ByName("27")
	busy := gpioreg.ByName("17")

	eeprom, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer eeprom.Close()

	o, err := inky.DetectOpts(eeprom)
	if err != nil {
		log.Printf("Here A")
		log.Fatal(err)
	}

	dev, err := inky.New(b, dc, reset, busy, o)
	if err != nil {
		log.Fatal(err)
	}
	return dev
}

func ClearDisplay() {
	dev := initialiseInkyWhat()

	if err := dev.Draw(img.Bounds(), img, image.Point{}); err != nil {
		log.Fatal(err)
	}

}

func SetImageOnInky(string path) {

	f, err := os.Open(*path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	dev := initialiseInkyWhat()

	log.Printf("Bounds %v", dev.Bounds())
	log.Printf("Bounds %v", img.Bounds())

	if err := dev.Draw(img.Bounds(), img, image.Point{}); err != nil {
		log.Fatal(err)
	}
}

func main() {
	path := flag.String("image", "", "Path to image file (400x300) to display")
	flag.Parse()

	if path != nil {
		SetImageOnInky(path)
	} else {
		ClearDisplay()
	}
}
