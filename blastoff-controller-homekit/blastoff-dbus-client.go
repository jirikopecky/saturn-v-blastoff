package main

import (
	"github.com/godbus/dbus/v5"
	"github.com/jirikopecky/saturn-v-blastoff/dbusutils"
)

// BlastOffClient represents DBus client of the BlastOff service
type BlastOffClient struct {
	conn     *dbus.Conn
	blastOff dbus.BusObject
}

// GetBlastOffClient creates new BlastOff client over given DBus connection
func GetBlastOffClient(conn *dbus.Conn) *BlastOffClient {
	obj := conn.Object(dbusutils.BlastOffDbusServiceName, dbusutils.BlastOffDbusObjectPath)

	return &BlastOffClient{
		conn:     conn,
		blastOff: obj,
	}
}

// Start starts the Blast Off animation over DBus
func (client *BlastOffClient) Start() error {
	result := client.blastOff.Call("Start", 0)
	return result.Err
}

// Stop stops the Blast Off animation over DBus
func (client *BlastOffClient) Stop() error {
	result := client.blastOff.Call("Stop", 0)
	return result.Err
}

// IsStarted gets the status of Blast Off animation via DBus
func (client *BlastOffClient) IsStarted() (bool, error) {
	var result bool
	err := client.blastOff.Call("IsStarted", 0).Store(&result)
	if err != nil {
		return false, err
	}

	return result, nil
}
