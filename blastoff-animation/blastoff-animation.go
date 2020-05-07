package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	log "github.com/sirupsen/logrus"
)

const (
	brightness = 60
	ledCounts  = 60
)

const (
	dbusObjectPath    dbus.ObjectPath = "/com/github/jirikopecky/SaturnV"
	dbusInterfaceName string          = "com.github.jirikopecky.SaturnV.BlastOff"
	dbusServiceName   string          = "com.github.jirikopecky.SaturnV"
)

const dbusIntrospect = `
<node>
	<interface name="` + dbusInterfaceName + `">
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
	app.dbus.Emit(dbusObjectPath, dbusInterfaceName+".StateChanged", true)

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
	app.dbus.Emit(dbusObjectPath, dbusInterfaceName+".StateChanged", true)

	return nil
}

func (app *animationApp) IsStarted() (bool, *dbus.Error) {
	app.mux.Lock()
	defer app.mux.Unlock()

	return (app.engine != nil), nil
}

func handleTermination(done chan bool) {
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// wait for signal
	sig := <-signals
	log.WithField("signal", sig).Debug("Signal received")

	done <- true
}

func privateSystemBus() (conn *dbus.Conn, err error) {
	conn, err = dbus.SystemBusPrivate()
	if err != nil {
		return nil, err
	}
	if err = conn.Auth(nil); err != nil {
		conn.Close()
		return nil, err
	}
	if err = conn.Hello(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil // success
}

func main() {
	conn, err := privateSystemBus()
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

	conn.Export(app, dbusObjectPath, dbusInterfaceName)
	conn.Export(introspect.Introspectable(dbusIntrospect),
		dbusObjectPath, "org.freedesktop.DBus.Introspectable")

	reply, err := conn.RequestName(dbusServiceName, dbus.NameFlagDoNotQueue)
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
