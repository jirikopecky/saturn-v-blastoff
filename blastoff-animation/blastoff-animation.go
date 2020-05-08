package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/jirikopecky/saturn-v-blastoff/dbusutils"
	log "github.com/sirupsen/logrus"
)

const (
	brightness = 60
	ledCounts  = 60
)

const dbusIntrospect = `
<node>
	<interface name="` + dbusutils.BlastOffDbusInterfaceName + `">
		<method name="Start" />
		<method name="Stop" />
		<method name="IsStarted">
			<arg name="isStarted" direction="out" type="b" />
		</method>
		<signal name="StateChanged">
			<arg name="isStarted" direction="out" type="b" />
		</signal>
	</interface>` + introspect.IntrospectDataString + `</node> `

type animationApp struct {
	engine *AnimationEngine
	mux    *sync.Mutex
	dbus   *dbus.Conn
}

func init() {
	log.SetLevel(log.TraceLevel)
}

func (app *animationApp) Start() *dbus.Error {
	app.mux.Lock()
	defer app.mux.Unlock()

	if app.engine != nil {
		log.Trace("Start called but animation already running.")
		return nil
	}

	engine, err := InitAnimationEngine(brightness, ledCounts)
	if err != nil {
		log.WithError(err).Error("Cannot initialize LED stripe driver!")
		return dbus.MakeFailedError(err)
	}

	err = engine.Setup()
	if err != nil {
		log.WithError(err).Error("Cannot initialize animation engine!")
		return dbus.MakeFailedError(err)
	}

	go engine.StartAnimation()

	app.engine = engine

	// emit DBus signal about the change
	app.dbus.Emit(dbusutils.BlastOffDbusObjectPath, dbusutils.BlastOffDbusInterfaceName+".StateChanged", true)

	return nil
}

func (app *animationApp) Stop() *dbus.Error {
	app.mux.Lock()
	defer app.mux.Unlock()

	if app.engine == nil {
		log.Trace("Stop called but animation is not running.")
		return nil
	}

	err := app.engine.CleanAndDestroy()
	if err != nil {
		log.WithError(err).Error("Error when cleaning up the animation!")
		return dbus.MakeFailedError(err)
	}

	app.engine = nil

	// emit DBus signal about the change
	app.dbus.Emit(dbusutils.BlastOffDbusObjectPath, dbusutils.BlastOffDbusInterfaceName+".StateChanged", true)

	return nil
}

func (app *animationApp) IsStarted() (bool, *dbus.Error) {
	app.mux.Lock()
	defer app.mux.Unlock()

	return (app.engine != nil), nil
}

func handleTermination(done chan<- bool) {
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// wait for signal
	sig := <-signals
	log.WithField("signal", sig).Debug("Signal received")

	done <- true
}

func main() {
	conn, err := dbusutils.ConnectPrivateSystemBus()
	if err != nil {
		log.WithError(err).Panic("Cannot connect to DBus System bus")
	}
	defer conn.Close()

	app := &animationApp{
		engine: nil,
		mux:    &sync.Mutex{},
		dbus:   conn,
	}
	defer app.Stop()

	done := make(chan bool, 1)
	go handleTermination(done)

	conn.Export(app, dbusutils.BlastOffDbusObjectPath, dbusutils.BlastOffDbusInterfaceName)
	conn.Export(introspect.Introspectable(dbusIntrospect), dbusutils.BlastOffDbusObjectPath, dbusutils.IntrospectableInterfaceName)

	reply, err := conn.RequestName(dbusutils.BlastOffDbusServiceName, dbus.NameFlagDoNotQueue)
	if err != nil {
		log.WithError(err).Panic("Failed to call DBus RequestName method")
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		log.Fatal("DBus name already taken")
	}

	log.Info("Started")
	<-done
	log.Info("Exiting")
}
