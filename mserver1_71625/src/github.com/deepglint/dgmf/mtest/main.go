package main

import (
	"os"

	ui "github.com/gizak/termui"
)

func main() {
	err := ui.Init()
	if err != nil {
		os.Exit(0)
	}
	defer ui.Close()

	ui.Body.Align()
	p := ui.NewPar("q: Quit MTool")
	p.Height = 3
	p.Width = 70
	p.TextFgColor = ui.ColorWhite
	p.BorderLabel = "Commands:"
	p.BorderFg = ui.ColorCyan

	g := ui.NewGauge()
	g.Percent = 50
	g.Width = 50
	g.Height = 3
	g.Y = 11
	g.BorderLabel = "Gauge"
	g.BarColor = ui.ColorRed
	g.BorderFg = ui.ColorWhite
	g.BorderLabelFg = ui.ColorCyan

	ui.Render(p, g) // feel free to call Render, it's async and non-block

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		// press q to quit
		ui.StopLoop()
	})

	ui.Handle("/timer/1s", func(e ui.Event) {
		// t := e.Data.(ui.EvtTimer)
		// // t is a EvtTimer
		// if t.Count%2 == 0 {
		// 	// do something
		// }
		g.Percent += 10
		ui.Render(p, g)
	})

	ui.Loop() // block until StopLoop is called
}

// package main

// import "github.com/gizak/termui"

// func main() {
// 	err := termui.Init()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer termui.Close()

// 	//termui.UseTheme("helloworld")

// 	data := []int{4, 2, 1, 6, 3, 9, 1, 4, 2, 15, 14, 9, 8, 6, 10, 13, 15, 12, 10, 5, 3, 6, 1, 7, 10, 10, 14, 13, 6}
// 	spl0 := termui.NewSparkline()
// 	spl0.Data = []int{100, 100, 100, 200}
// 	spl0.Title = "Sparkline 0"
// 	spl0.LineColor = termui.ColorGreen

// 	// single
// 	spls0 := termui.NewSparklines(spl0)
// 	spls0.Height = 2
// 	spls0.Width = 20
// 	spls0.Border = false

// 	spl1 := termui.NewSparkline()
// 	spl1.Data = []int{100, 100, 100, 200}
// 	spl1.Title = "Sparkline 1"
// 	spl1.LineColor = termui.ColorRed

// 	spl2 := termui.NewSparkline()
// 	spl2.Data = data[5:]
// 	spl2.Title = "Sparkline 2"
// 	spl2.LineColor = termui.ColorMagenta

// 	// group
// 	spls1 := termui.NewSparklines(spl0, spl1, spl2)
// 	spls1.Height = 8
// 	spls1.Width = 20
// 	spls1.Y = 3
// 	spls1.BorderLabel = "Group Sparklines"

// 	spl3 := termui.NewSparkline()
// 	spl3.Data = []int{100, 100, 150, 100, 300}
// 	spl3.Title = "Enlarged Sparkline"
// 	spl3.Height = 8
// 	spl3.LineColor = termui.ColorYellow

// 	spls2 := termui.NewSparklines(spl3)
// 	spls2.Height = 11
// 	spls2.Width = 30
// 	spls2.BorderFg = termui.ColorCyan
// 	spls2.X = 21
// 	spls2.BorderLabel = "Tweeked Sparkline"

// 	termui.Render(spls0, spls1, spls2)

// 	termui.Handle("/sys/kbd/q", func(termui.Event) {
// 		termui.StopLoop()
// 	})
// 	termui.Loop()

// }