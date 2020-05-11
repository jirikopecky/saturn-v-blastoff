package main

import (
	"flag"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/jirikopecky/saturn-v-blastoff/dbusutils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func switchState(action string) {
	log.WithField("Action", action).Info("State change")
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "./config.toml", "Path to configuration file")

	flag.Parse()

	conn, err := dbusutils.ConnectPrivateSystemBus()
	if err != nil {
		log.WithError(err).Panic("Cannot connect to DBus System bus")
	}
	defer conn.Close()

	blastOffClient := GetBlastOffClient(conn)

	viper.SetDefault("pin", "05042020")
	viper.SetDefault("database", "./db")
	viper.SetDefault("port", "")

	viper.SetConfigType("toml")
	viper.SetConfigFile(configPath)
	if err = viper.ReadInConfig(); err != nil {
		log.WithError(err).Panic("Cannot read config file")
	}

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
		Pin:         viper.GetString("pin"),
		Port:        viper.GetString("port"),
		StoragePath: viper.GetString("database"),
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
