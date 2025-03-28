package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/godbus/dbus/v5"
)

type AccessPoint struct {
	Path     dbus.ObjectPath
	SSID     string
	Strength uint8
	Flags    uint32
}

type SavedNetwork struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
}

func getWifiDevicePath(conn *dbus.Conn) (dbus.ObjectPath, error) {
	obj := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")

	var devicePaths []dbus.ObjectPath
	err := obj.Call("org.freedesktop.NetworkManager.GetAllDevices", 0).Store(&devicePaths)

	if err != nil {
		return "", fmt.Errorf("GetAllDevices failed: %w", err)
	}

	for _, path := range devicePaths {
		deviceObj := conn.Object("org.freedesktop.NetworkManager", path)
		variant, err := deviceObj.GetProperty("org.freedesktop.NetworkManager.Device.DeviceType")
		if err != nil {
			return "", fmt.Errorf("GetProperty failed: %w", err)
		}

		deviceType, ok := variant.Value().(uint32)
		if !ok {
			return "", fmt.Errorf("type assertion failed. DeviceType is not uint32")
		}

		if deviceType == 2 {
			return path, nil
		}
	}

	return "", fmt.Errorf("No wifi device found.")
}

func getAccessPoints(conn *dbus.Conn, wifiDevicePath dbus.ObjectPath) ([]AccessPoint, error) {
	wifiDeviceObj := conn.Object("org.freedesktop.NetworkManager", wifiDevicePath)

	var accessPointsPaths []dbus.ObjectPath
	err := wifiDeviceObj.Call("org.freedesktop.NetworkManager.Device.Wireless.GetAllAccessPoints", 0).Store(&accessPointsPaths)
	if err != nil {
		return nil, fmt.Errorf("GetAllAccessPoints failed: %w", err)
	}

	var accessPoints []AccessPoint
	for _, apPath := range accessPointsPaths {
		apObj := conn.Object("org.freedesktop.NetworkManager", apPath)
		ssidVariant, err := apObj.GetProperty("org.freedesktop.NetworkManager.AccessPoint.Ssid")
		if err != nil {
			fmt.Printf("Failed to get SSID: %s\n", err)
			continue
		}

		ssidBytes, ok := ssidVariant.Value().([]byte)
		if !ok {
			fmt.Println("type assertion failed. SSID is not []byte")
			continue
		}
		ssid := string(ssidBytes)

		strengthVariant, err := apObj.GetProperty("org.freedesktop.NetworkManager.AccessPoint.Strength")
		if err != nil {
			fmt.Printf("failed to get Strength: %s\n", err)
			continue
		}

		strength, ok := strengthVariant.Value().(uint8)
		if !ok {
			fmt.Println("type assertion failed. Strength is not uint8")
			continue
		}

		flagsVariant, err := apObj.GetProperty("org.freedesktop.NetworkManager.AccessPoint.Flags")
		if err != nil {
			fmt.Printf("failed tp get Flags: %s\n", err)
			continue
		}

		flags, ok := flagsVariant.Value().(uint32)
		if !ok {
			fmt.Println("type assertion failed. Flags is not uint32")
			continue
		}

		accessPoints = append(accessPoints, AccessPoint{
			Path:     apPath,
			SSID:     ssid,
			Strength: strength,
			Flags:    flags,
		})
	}

	return accessPoints, nil
}

func createConnectionSettings(ap AccessPoint, password string) map[string]map[string]dbus.Variant {
	settings := map[string]map[string]dbus.Variant{
		"802-11-wireless": {
			"mode":     dbus.MakeVariant("infrastructure"),
			"ssid":     dbus.MakeVariant([]byte(ap.SSID)),
			"security": dbus.MakeVariant("802-11-wireless-security"),
		},
		"connection": {
			"id":          dbus.MakeVariant(ap.SSID),
			"type":        dbus.MakeVariant("802-11-wireless"),
			"autoconnect": dbus.MakeVariant(true),
		},
		"ipv4": {
			"method": dbus.MakeVariant("auto"),
		},
		"ipv6": {
			"method": dbus.MakeVariant("auto"),
		},
	}

	// Add password if needed
	if (ap.Flags & 0x1) > 0 {
		settings["802-11-wireless-security"] = map[string]dbus.Variant{
			"key-mgmt": dbus.MakeVariant("wpa-psk"),
			"psk":      dbus.MakeVariant(password),
		}
	} else {
		// If the network is not encrypted, don't include security settings
		// No additional settings needer here
	}

	return settings
}

func connectToNetwork(conn *dbus.Conn, settings map[string]map[string]dbus.Variant, wifiDevicePath dbus.ObjectPath, apPath dbus.ObjectPath) (dbus.ObjectPath, error) {
	obj := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")

	var newConnectionPath, activeConnectionPath dbus.ObjectPath
	err := obj.Call("org.freedesktop.NetworkManager.AddAndActivateConnection", 0, settings, wifiDevicePath, apPath).Store(&newConnectionPath, &activeConnectionPath)
	if err != nil {
		return "", fmt.Errorf("AddAndActivateConnection failed: %w", err)
	}

	return activeConnectionPath, nil
}

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

	if state == 1 {
		return true, nil
	}

	return false, nil
}

func waitForConnection(timeoutSeconds int) {
	action := func() {
		time.Sleep(time.Duration(timeoutSeconds) * time.Second)
	}
	if err := spinner.New().Title("Establishing connection...").Action(action).Run(); err != nil {
		fmt.Println("Spinner establish failed:", err)
	}
}

func waitForScan(timeoutSeconds int) {
	action := func() {
		time.Sleep(time.Duration(timeoutSeconds) * time.Second)
	}
	if err := spinner.New().Title("Scanning the nearby networks...").Action(action).Run(); err != nil {
		fmt.Println("Spinner scan failed:", err)
	}
}

func forceWifiScan(conn *dbus.Conn, wifiDevicePath dbus.ObjectPath) error {
	deviceObj := conn.Object("org.freedesktop.NetworkManager", wifiDevicePath)
	options := map[string]dbus.Variant{}
	err := deviceObj.Call("org.freedesktop.NetworkManager.Device.Wireless.RequestScan", 0, options).Store()
	return err
}

func savePassword(ssid, password string) error {
	var savedNetworks []SavedNetwork

	data, err := os.ReadFile("saved_networks.json")
	if err == nil {
		json.Unmarshal(data, &savedNetworks)
	}

	for i, network := range savedNetworks {
		if network.SSID == ssid {
			savedNetworks[i].Password = password
			file, err := json.MarshalIndent(savedNetworks, "", " ")
			if err != nil {
				return err
			}
			return os.WriteFile("saved_networks.json", file, 0644)
		}
	}

	savedNetworks = append(savedNetworks, SavedNetwork{SSID: ssid, Password: password})

	file, err := json.MarshalIndent(savedNetworks, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile("saved_networks.json", file, 0644)
}

func loadPassword(ssid string) (string, bool) {
	data, err := os.ReadFile("saved_networks.json")
	if err != nil {
		return "", false
	}

	var saved_networks []SavedNetwork
	json.Unmarshal(data, &saved_networks)

	for _, network := range saved_networks {
		if network.SSID == ssid {
			return network.Password, true
		}
	}

	return "", false
}

func turnOnWifi() error {
	cmd := exec.Command("nmcli", "radio", "wifi", "on")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to enable wifi: %w, output: %s", err, output)
	}

	return nil
}

func turnOffWifi() error {
	cmd := exec.Command("nmcli", "radio", "wifi", "off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to enable wifi: %w, output: %s", err, output)
	}

	return nil
}

func showHelpMessage() {
	fmt.Println("Usage: ewc [Option]")
	fmt.Println("Options:")
	fmt.Println(" on   turns on the wifi")
	fmt.Println(" off  turns off the wifi")
	fmt.Println(" help shows this message")
}

func main() {

	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Printf("SystemBus failed: %s\n", err)
		os.Exit(1)
	}

	// Collect args for turning on and off the wifi
	args := os.Args[1:]
	if !slices.Equal(args, nil) {
		if args[0] == "on" {
			if err := turnOnWifi(); err != nil {
				fmt.Println(err)
			}
			fmt.Println("Wi-Fi Enabled.")
			os.Exit(0)
		} else if args[0] == "off" {
			if err := turnOffWifi(); err != nil {
				fmt.Println(err)
			}
			fmt.Println("Wi-Fi Disabled.")
			os.Exit(0)
		} else if args[0] == "help" {
			showHelpMessage()
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
