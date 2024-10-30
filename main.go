package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var progress float32
var progressIncrementer chan float32

func main() {
	// Setup a channel to provide ticks to increment progress
	progressIncrementer = make(chan float32)
	go func() {
		for {
			time.Sleep(time.Second / 25)
			progressIncrementer <- 0.004
		}
	}()

	go func() {
		// Create new window
		w := new(app.Window)
		w.Option(app.Title("Egg timer"))
		w.Option(app.Size(unit.Dp(400), unit.Dp(600)))
		if err := draw(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

type C = layout.Context
type D = layout.Dimensions

func draw(w *app.Window) error {

	// ops are the operations form the UI
	var ops op.Ops

	// startButton is a clickable widget
	var startButton widget.Clickable

	// boilDurationInput is a textfield to input boil duration
	var boilDurationInput widget.Editor

	// is the egg boiling?
	var boiling bool
	var boilDuration float32

	// theme defines the material design style
	theme := material.NewTheme()

	go func() {
		for range progressIncrementer {
			if boiling && progress < 1 {
				progress += 1.0 / 25.0 / boilDuration

				if progress >= 1 {
					progress = 1
				}

				w.Invalidate()
			}
		}
	}()

	// Listen for events in the window
	for {
		// First, grab the event

		// Then, detect the type
		switch evt := w.Event().(type) {

		// This is sent when the app should re-render
		case app.FrameEvent:
			gtx := app.NewContext(&ops, evt)

			if startButton.Clicked(gtx) {
				// Start or stop the timer
				boiling = !boiling

				// Reset the boil
				if progress >= 1 {
					progress = 0
				}

				inputString := boilDurationInput.Text()
				inputString = strings.TrimSpace(inputString)
				inputFloat, _ := strconv.ParseFloat(inputString, 32)
				boilDuration = float32(inputFloat)
				boilDuration = boilDuration / (1 - progress)
			}
			layout.Flex{
				// Vertical alignment, from top to bottom
				Axis: layout.Vertical,
				// Empty space is left at the start, i.e. at the top
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				layout.Rigid(
					func(gtx C) D {
						// Draw a custom path, shaped like an egg
						var eggPath clip.Path
						op.Offset(image.Pt(gtx.Dp(200), gtx.Dp(25))).Add(gtx.Ops)
						eggPath.Begin(gtx.Ops)
						// Rotate from 0 to 360 degrees
						for deg := 0.0; deg <= 360; deg++ {

							// Egg math (really) at this brilliant site. Thanks!
							// https://observablehq.com/@toja/egg-curve
							// Convert degrees to radians
							rad := deg / 360 * 2 * math.Pi
							// Trig gives the distance in X and Y direction
							cosT := math.Cos(rad)
							sinT := math.Sin(rad)
							// Constants to define the eggshape
							a := 110.0
							b := 150.0
							d := 20.0
							// The x/y coordinates
							x := a * cosT
							y := -(math.Sqrt(b*b-d*d*cosT*cosT) + d*sinT) * sinT
							// Finally the point on the outline
							p := f32.Pt(float32(x), float32(y))
							// Draw the line to this point
							eggPath.LineTo(p)
						}
						// Close the path
						eggPath.Close()

						// Get hold of the actual clip
						eggArea := clip.Outline{Path: eggPath.End()}.Op()

						// Fill the shape
						// color := color.NRGBA{R: 255, G: 239, B: 174, A: 255}
						color := color.NRGBA{R: 255, G: uint8(239 * (1 - progress)), B: uint8(174 * (1 - progress)), A: 255}
						paint.FillShape(gtx.Ops, color, eggArea)

						d := image.Point{Y: 375}
						return layout.Dimensions{Size: d}
					},
				),
				layout.Rigid(
					func(gtx C) D {
						// Wrap the editor in material design
						ed := material.Editor(theme, &boilDurationInput, "sec")

						// Define characteristics of the input box
						boilDurationInput.SingleLine = true
						boilDurationInput.Alignment = text.Middle

						// Count down the text when boiling
						if boiling && progress < 1 {
							boilRemain := (1 - progress) * boilDuration
							inputStr := fmt.Sprintf("%.1f", math.Round(float64(boilRemain)*10)/10)
							boilDurationInput.SetText(inputStr)
						}

						// Define insets
						margins := layout.Inset{
							Top:    unit.Dp(0),
							Right:  unit.Dp(170),
							Bottom: unit.Dp(40),
							Left:   unit.Dp(179),
						}

						// Define borders
						border := widget.Border{
							Color:        color.NRGBA{R: 204, G: 204, B: 204, A: 255},
							CornerRadius: unit.Dp(3),
							Width:        unit.Dp(2),
						}

						// Layout borders inside of margins
						return margins.Layout(gtx,
							func(gtx C) D {
								return border.Layout(gtx, ed.Layout)
							},
						)

					},
				),
				layout.Rigid(
					func(gtx C) D {
						bar := material.ProgressBar(theme, progress)
						return bar.Layout(gtx)
					},
				),
				layout.Rigid(
					func(gtx C) D {
						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Left:   unit.Dp(25),
							Right:  unit.Dp(25),
						}

						return margins.Layout(gtx,
							func(gtx C) D {
								text := "Start"
								if boiling && progress < 1 {
									text = "Stop"
								} else {
									text = "Finished"
								}
								btn := material.Button(theme, &startButton, text)
								return btn.Layout(gtx)
							})
					},
				),
				layout.Rigid(
					layout.Spacer{Height: unit.Dp(25)}.Layout,
				),
			)
			evt.Frame(gtx.Ops)

		// And this is sent when the application should close
		case app.DestroyEvent:
			return evt.Err
		}
	}
}
