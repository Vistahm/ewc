package main

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

// Creates a d-bus object and activates a connection
func connectToNetwork(conn *dbus.Conn, settings map[string]map[string]dbus.Variant, wifiDevicePath dbus.ObjectPath, apPath dbus.ObjectPath) (dbus.ObjectPath, error) {
	obj := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")

	var newConnectionPath, activeConnectionPath dbus.ObjectPath
	err := obj.Call("org.freedesktop.NetworkManager.AddAndActivateConnection", 0, settings, wifiDevicePath, apPath).Store(&newConnectionPath, &activeConnectionPath)
	if err != nil {
		return "", fmt.Errorf("AddAndActivateConnection failed: %w", err)
	}

	return activeConnectionPath, nil
}

// Checks to see if the connection attempt was successful or not
func checkConnectionState(conn *dbus.Conn, activeConnectionPath dbus.ObjectPath) (bool, error) {
	activeConnectionObj := conn.Object("org.freedesktop.NetworkManager", activeConnectionPath)

	activeVariant, err := activeConnectionObj.GetProperty("org.freedesktop.NetworkManager.Connection.Active.State")
	state, ok := activeVariant.Value().(uint32)
	if !ok {
		return false, fmt.Errorf("State is not uint32: %w", err)
	}
	fmt.Printf("State: %d\n", state)

	if err != nil {
		return false, fmt.Errorf("failed to get connection state: %w", err)
	}

	if state == 1 || state == 2 {
		return true, nil
	}

	return false, nil
}

func forceWifiScan(conn *dbus.Conn, wifiDevicePath dbus.ObjectPath) error {
	deviceObj := conn.Object("org.freedesktop.NetworkManager", wifiDevicePath)
	options := map[string]dbus.Variant{}
	err := deviceObj.Call("org.freedesktop.NetworkManager.Device.Wireless.RequestScan", 0, options).Store()
	return err
}
