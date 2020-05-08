package main

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	hcLog "github.com/brutella/hc/log"
	"github.com/jirikopecky/saturn-v-blastoff/dbusutils"
	log "github.com/sirupsen/logrus"
)

func switchState(action string) {
	log.WithField("Action", action).Info("State change")
}

func main() {
	hcLog.Debug.Enable()

	conn, err := dbusutils.ConnectPrivateSystemBus()
	if err != nil {
		log.WithError(err).Panic("Cannot connect to DBus System bus")
	}
	defer conn.Close()

	blastOffClient := GetBlastOffClient(conn)

	// TODO: take these from config
	pin := "05042020"
	port := ""
	database := "./db"

	info := accessory.Info{
		Name:         "Saturn V",
		Manufacturer: "Jiri Kopecky",
		Model:        "Saturn V Stand",
		SerialNumber: "1",
	}

	acc := accessory.NewSwitch(info)

	acc.Switch.On.OnValueRemoteUpdate(func(on bool) {
		log.WithField("Status", on).Trace("State change remote requested")

		var err error
		if on == true {
			err = blastOffClient.Start()
		} else {
			err = blastOffClient.Stop()
		}

		if err != nil {
			log.WithError(err).Panic("Call to DBus service failed")
		}

		log.WithField("Status", on).Trace("State change remote updated")
	})

	acc.Switch.On.OnValueRemoteGet(func() bool {
		log.Trace("State read request")

		isStarted, err := blastOffClient.IsStarted()
		if err != nil {
			log.WithError(err).Panic("Call to DBus service failed")
		}

		log.WithField("Status", isStarted).Trace("State read finished")

		return isStarted
	})

	hcConfig := hc.Config{
		Pin:         pin,
		Port:        port,
		StoragePath: database,
	}
	t, err := hc.NewIPTransport(hcConfig, acc.Accessory)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})

	t.Start()
}
