package main

import (
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
)

var savedNetworksFile = getConfigFilePath()

func main() {

	// Create a system bus
	conn, err := dbus.SystemBus()
	handleError(err, "SystemBus failed")

	// Collect the arguments and handle them
	args := os.Args[1:]
	handleArguments(args)

	obj := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")

	// NetworkManager state
	getNetworkManagerState(obj)

	wifiDevicePath, err := getWifiDevicePath(conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Force scan for available networks
	if err := forceWifiScan(conn, wifiDevicePath); err != nil {
		fmt.Println("failed to initiate wifi scan:", err)
		os.Exit(1)
	}

	// Wait to scan all the access points
	waitForScan(10)

	accessPoints, err := getAccessPoints(conn, wifiDevicePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Selecting an access point
	selectedAP := selectAccessPoint(accessPoints)
	password := getPasswordForAccessPoint(selectedAP)

	settings := createConnectionSettings(selectedAP, password)

	activeConnectionPath, err := connectToNetwork(conn, settings, wifiDevicePath, selectedAP.Path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Wait
	waitForConnection(5)

	connected, err := checkConnectionState(conn, activeConnectionPath)
	if err != nil {
		fmt.Println("failed to verify connection:", err)
		os.Exit(1)
	}

	if connected {
		fmt.Printf("Successfully connected to: %s\n", selectedAP.SSID)

		// Save the password for ssid
		if password != "" {
			if err := savePassword(selectedAP.SSID, password); err != nil {
				fmt.Println("failed to save password:", err)
			}
		}
	} else {
		fmt.Println("Failed to connect to the network. Please check your password.")
		os.Exit(1)
	}

}
