package app

import (
	"fmt"

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
