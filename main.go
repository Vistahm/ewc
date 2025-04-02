package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/godbus/dbus/v5"
)

var savedNetworksFile = getConfigFilePath()

func main() {

	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Printf("SystemBus failed: %s\n", err)
		os.Exit(1)
	}

	// Collect args for turning on and off the wifi
	args := os.Args[1:]

	if !slices.Equal(args, nil) {
		switch args[0] {
		case "on":
			if err := turnOnWifi(); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Wi-Fi Enabled.")
			}
			os.Exit(0)

		case "off":
			if err := turnOffWifi(); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Wi-Fi Disabled.")
			}
			os.Exit(0)

		case "forget":
			if len(args) < 2 {
				fmt.Println("Please provide an SSID to forget.")
				os.Exit(1)
			}

			ssidToForget := args[1]
			if err := forgetNetwork(ssidToForget); err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("Success.\nForgotten network: %s\n", ssidToForget)
			}
			os.Exit(0)

		case "help":
			showHelpMessage()
			os.Exit(0)

		default:
			fmt.Println("Unknown command. Use 'help' for a list of commands.")
			os.Exit(0)
		}
	}

	obj := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")

	var state uint32
	err = obj.Call("org.freedesktop.NetworkManager.state", 0).Store(&state)
	if err != nil {
		fmt.Printf("Call to state failed: %s\n", err)
		os.Exit(1)
	}

	switch state {
	case 70:
		fmt.Println("NetworkManager is connected globally.")
	case 60:
		fmt.Println("NetworkManager is connected to a local Network.")
	case 50:
		fmt.Println("NetworkManager is connecting.")
	case 40:
		fmt.Println("NetworkManager is disconnected.")
	case 20:
		fmt.Println("NetworkManager is sleeping.")
	case 10:
		fmt.Println("NetworkManager is unavailable.")
	case 0:
		fmt.Println("NetworkManager's status is unknown.")
	default:
		fmt.Println("unknown NetworkManager state.")
	}

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

	// huh forms
	// 1. select access point
	var selectedAP AccessPoint
	var ssidOptions []huh.Option[AccessPoint]

	for _, ap := range accessPoints {
		ssidOptions = append(ssidOptions, huh.NewOption(fmt.Sprintf("%s (Strength: %d)", ap.SSID, ap.Strength), ap))
	}

	selectForm := huh.NewSelect[AccessPoint]().
		Title("Select Wi-Fi Network").
		Options(ssidOptions...).
		Value(&selectedAP)

	err = selectForm.Run()
	if err != nil {
		fmt.Println("error with form:", err)
		os.Exit(1)
	}

	// load saved password for the selected ssid
	var password string
	savedPassword, found := loadPassword(selectedAP.SSID)

	if (selectedAP.Flags & 0x1) > 0 {
		if found {
			// if saved password is found, skip the password prompt
			password = savedPassword
			fmt.Println("Using saved password for:", selectedAP.SSID)
		} else {
			// if no password saved, prompt the user
			var passwordInput string
			passwordForm := huh.NewInput().
				Title("Enter password:").
				EchoMode(huh.EchoModePassword).
				Value(&passwordInput)

			form := huh.NewForm(
				huh.NewGroup(passwordForm),
			)
			err = form.Run()
			if err != nil {
				fmt.Println("Error with password form:", err)
				os.Exit(1)
			}

			password = passwordInput
		}
	} else {
		fmt.Println("No password required for this network.")
	}

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
