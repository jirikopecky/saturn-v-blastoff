package dbusutils

import "github.com/godbus/dbus/v5"

// ConnectPrivateSystemBus creates non-shared connection to DBus System Bus
func ConnectPrivateSystemBus() (conn *dbus.Conn, err error) {
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
