package app

import (
	"fmt"
	"os"
	"time"

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

func GetWifiDevicePath(conn *dbus.Conn) (dbus.ObjectPath, error) {
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

func GetAccessPoints(conn *dbus.Conn, wifiDevicePath dbus.ObjectPath) ([]AccessPoint, error) {
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

func CreateConnectionSettings(ap AccessPoint, password string) map[string]map[string]dbus.Variant {
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
		"802-11-wireless-security": {
			"key-mgmt": dbus.MakeVariant("wpa-psk"),
			"psk":      dbus.MakeVariant(password),
		},
		"ipv4": {
			"method": dbus.MakeVariant("auto"),
		},
		"ipv6": {
			"method": dbus.MakeVariant("auto"),
		},
	}

	// Add password if needed
	//if ap.Flags == 0 || (ap.Flags&0x1) > 0 {
	//	settings["802-11-wireless-security"] = map[string]dbus.Variant{
	//		"key-mgmt": dbus.MakeVariant("wpa-psk"),
	//		"psk":      dbus.MakeVariant(password),
	//	}
	//} else {
	//	// If the network is not encrypted, don't include security settings
	//	// No additional settings needer here
	//}

	// remove security settings if the network is unencrypted
	if password == "" {
		delete(settings, "802-11-wireless-security")
		wirelessSettings := settings["802-11-wireless"]
		delete(wirelessSettings, "security")
		settings["802-11-wireless"] = wirelessSettings
	}

	return settings
}

func ConnectToSSID(ssid string) {
	// print the hint
	fmt.Println(Yellow+"   You're using direct connection and the program doesn't scan your nearby networks.", "\n",
		"   So you should be aware of the SSID that you're trying to connect to."+Reset+"\n")

	conn, err := dbus.SystemBus()
	HandleError(err, "SystemBus failed (ConnectToSSID)")

	wifiDevicePath, err := GetWifiDevicePath(conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// laod saved password
	password, found := LoadPassword(ssid)

	var settings map[string]map[string]dbus.Variant
	if found {
		// Load password if already saved
		fmt.Println("Using saved password for:", ssid)
		ap := AccessPoint{SSID: ssid}
		settings = CreateConnectionSettings(ap, password)
	} else {
		// Prompt the user for password if not found
		password = PromptForPassword()
		ap := AccessPoint{SSID: ssid}
		settings = CreateConnectionSettings(ap, password)
	}

	// if no password entered, treat the network as unencrypted
	if password == "" {
		fmt.Println(Cyan + "No password entered. The selected network will be treated as unencrypted." + Reset)
		ap := AccessPoint{SSID: ssid}
		settings = CreateConnectionSettings(ap, password)
	}

	activeConnectionPath, err := ConnectToNetwork(conn, settings, wifiDevicePath, "/")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	WaitForConnection(5)

	// retry mechanism
	timeout := 5 * time.Second
	startTime := time.Now()
	connected := false

	for time.Since(startTime) < timeout {
		connected, err = CheckConnectionState(conn, activeConnectionPath)
		if err != nil {
			// consider logging the error or taking other actions
			time.Sleep(1 * time.Second)
			continue
		}

		if connected {
			break // Successfull connection
		}
	}

	if connected {
		fmt.Println("Successfully connected to:", ssid)

		// save password
		if password != "" {
			if err := SavePassword(ssid, password); err != nil {
				fmt.Println("failed to save password:", err)
			}
		}
	} else {
		fmt.Println("Connection was not established. Wrong password maybe?")
		os.Exit(1)
	}

}
