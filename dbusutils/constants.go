package dbusutils

import "github.com/godbus/dbus/v5"

const (
	// BlastOffDbusObjectPath is DBus Object path to BlastOff interface
	BlastOffDbusObjectPath dbus.ObjectPath = "/com/github/jirikopecky/SaturnV"

	// BlastOffDbusInterfaceName is DBus interface name of BlastOff service
	BlastOffDbusInterfaceName string = "com.github.jirikopecky.SaturnV.BlastOff"

	// BlastOffDbusServiceName is DBus service name used by the BlastOff animation service
	BlastOffDbusServiceName string = "com.github.jirikopecky.SaturnV"

	// IntrospectableInterfaceName is the name of DBus Introspectable interface - org.freedesktop.DBus.Introspectable
	IntrospectableInterfaceName string = "org.freedesktop.DBus.Introspectable"
)
