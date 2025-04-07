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
	//dev := initialiseInkyWhat()
	//
	//if err := dev.Draw(image.Rectangle{image.Point{0, 0}, image.Point{400, 300}}, img, image.Point{}); err != nil {
	//	log.Fatal(err)
	//}
}

func SetImageOnInky(path *string) {

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

// randomPoints returns some random x, y points.
//func randomPoints(n int) plotter.XYs {
//	pts := make(plotter.XYs, n)
//	for i := range pts {
//		if i == 0 {
//			pts[i].X = rand.Float64()
//		} else {
//			pts[i].X = pts[i-1].X + rand.Float64()
//		}
//		pts[i].Y = pts[i].X + 10*rand.Float64()
//	}
//	return pts
//}

//func genImage() {
//	p := plot.New()
//
//	p.Title.Text = "Plotutil example"
//	p.X.Label.Text = "X"
//	p.Y.Label.Text = "Y"
//
//	err := plotutil.AddLinePoints(p,
//		"First", randomPoints(15),
//		"Second", randomPoints(15),
//		"Third", randomPoints(15))
//	if err != nil {
//		panic(err)
//	}
//
//	// Save the plot to a PNG file.
//	//if err := p.Save(vg.Points(300), vg.Points(225), "points.png"); err != nil {
//	//	panic(err)
//	//}
//	if err := betterSave(p, vg.Points(300), vg.Points(225), "points.png"); err != nil {
//		panic(err)
//	}
//
//	time.Sleep(2 * time.Second)
//	log.Printf("Saved plot to points.png")
//
//}

//func betterSave(p *plot.Plot, w vg.Length, h vg.Length, file string) error {
//	f, err := os.Create(file)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		e := f.Close()
//		if err == nil {
//			err = e
//		}
//	}()
//
//	format := strings.ToLower(filepath.Ext(file))
//	if len(format) != 0 {
//		format = format[1:]
//	}
//	c, err := p.WriterTo(w, h, format)
//	if err != nil {
//		return err
//	}
//
//	_, err = c.WriteTo(f)
//	f.Sync()
//	return err
//}

func ptr(s string) *string { return &s }

func main() {
	path := flag.String("image", "", "Path to image file (400x300) to display")
	flag.Parse()

	//path := strings.Clone("points.png")

	//genImage()
	//SetImageOnInky(ptr("prdoch.png"))
	if path != nil {
		SetImageOnInky(path)
	} else {
		ClearDisplay()
	}
}
