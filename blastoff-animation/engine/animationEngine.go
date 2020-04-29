// Package engine implements the animation of fire for LEGO Saturn V rocket stand.
package engine

import (
	"fmt"
	"time"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
	log "github.com/sirupsen/logrus"
)

const (
	sleepTime = 50
)

type wsEngine interface {
	Init() error
	Render() error
	Wait() error
	Fini()
	Leds(channel int) []uint32
}

// AnimationEngine responsible for running blast-off animation
type AnimationEngine struct {
	brightness int
	ledCount   int
	quit       chan struct{}
	ws         wsEngine
}

// InitAnimationEngine initializes and returns AnimationEngine
func InitAnimationEngine(brightness int, ledCount int) (*AnimationEngine, error) {
	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledCount

	dev, err := ws2811.MakeWS2811(&opt)
	if err != nil {
		return nil, err
	}

	engine := &AnimationEngine{
		brightness: brightness,
		ledCount:   ledCount,
		ws:         dev,
		quit:       make(chan struct{}),
	}

	log.WithFields(log.Fields{
		"brightness": brightness,
		"ledCount":   ledCount,
	}).Debug("Animation initialized")

	return engine, nil
}

// Setup sets up WS2811 LEDs
func (engine *AnimationEngine) Setup() error {
	return engine.ws.Init()
}

// CleanAndDestroy clears LEDs and release driver resources
func (engine *AnimationEngine) CleanAndDestroy() error {
	log.Trace("Closing animation engine")

	close(engine.quit)

	color := uint32(0x000000)
	for i := 0; i < len(engine.ws.Leds(0)); i++ {
		engine.ws.Leds(0)[i] = color
	}

	if err := engine.ws.Render(); err != nil {
		return err
	}

	engine.ws.Fini()

	log.Debug("Animation engine destroyed")

	return nil
}

func (engine *AnimationEngine) doAnimationStep(color uint32) error {
	log.WithField("color", fmt.Sprintf("%06X", color)).Trace("Doing animation step")

	for i := 0; i < len(engine.ws.Leds(0)); i++ {
		engine.ws.Leds(0)[i] = color
		if err := engine.ws.Render(); err != nil {
			return err
		}
		time.Sleep(sleepTime * time.Millisecond)
	}

	return nil
}

// StartAnimation starts the animation loop
func (engine *AnimationEngine) StartAnimation() error {
	colors := [3]uint32{0x0000ff, 0x00ff00, 0xff0000}
	colorIndex := 0

	for {
		select {
		case <-engine.quit:
			return nil
		default:
			err := engine.doAnimationStep(colors[colorIndex%len(colors)])
			if err != nil {
				return err
			}

			colorIndex++
		}
	}
}
