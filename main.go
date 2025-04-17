package main

import (
	"fmt"
	"os"

	"github.com/Vistahm/ewc/app"
	"github.com/godbus/dbus/v5"
)

func main() {

	// Create a system bus
	conn, err := dbus.SystemBus()
	app.HandleError(err, "SystemBus failed")

	// Collect the arguments and handle them
	args := os.Args[1:]
	app.HandleArguments(args)

	obj := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")

	// NetworkManager state
	app.GetNetworkManagerState(obj)

	wifiDevicePath, err := app.GetWifiDevicePath(conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Force scan for available networks
	if err := app.ForceWifiScan(conn, wifiDevicePath); err != nil {
		fmt.Println("failed to initiate wifi scan:", err)
		os.Exit(1)
	}

	// Wait to scan all the access points
	app.WaitForScan(10)

	accessPoints, err := app.GetAccessPoints(conn, wifiDevicePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Selecting an access point
	selectedAP := app.SelectAccessPoint(accessPoints)
	password := app.GetPasswordForAccessPoint(selectedAP)

	settings := app.CreateConnectionSettings(selectedAP, password)

	activeConnectionPath, err := app.ConnectToNetwork(conn, settings, wifiDevicePath, selectedAP.Path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Wait
	app.WaitForConnection(6)

	connected, err := app.CheckConnectionState(conn, activeConnectionPath)
	if err != nil {
		fmt.Println("failed to verify connection: please check your password again")
		os.Exit(1)
	}

	if connected {
		fmt.Printf("Successfully connected to: %s\n", selectedAP.SSID)

		// Save the password for ssid
		if password != "" {
			if err := app.SavePassword(selectedAP.SSID, password); err != nil {
				fmt.Println("failed to save password:", err)
			}
		}
	} else {
		fmt.Println("Failed to connect to the network. Please check your password.")
		os.Exit(1)
	}

}
