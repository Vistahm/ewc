package app

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
)

var savedNetworksFile = GetConfigFilePath()

// Saves the password of the ssid if the connection was successful
func SavePassword(ssid, password string) error {
	var savedNetworks []SavedNetwork

	data, err := os.ReadFile(savedNetworksFile)
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
			return os.WriteFile(savedNetworksFile, file, 0644)
		}
	}

	savedNetworks = append(savedNetworks, SavedNetwork{SSID: ssid, Password: password})

	file, err := json.MarshalIndent(savedNetworks, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(savedNetworksFile, file, 0644)
}

// Loads the password of ssid
func LoadPassword(ssid string) (string, bool) {
	data, err := os.ReadFile(savedNetworksFile)
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

// Specifies the full path to save the json file in the user's home directory
func GetConfigFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Unable to find the home directory")
	}

	return filepath.Join(homeDir, ".saved_networks.json")
}

// Forgets the network based on the ssid
func ForgetNetwork(ssid string) error {
	data, err := os.ReadFile(savedNetworksFile)
	if err != nil {
		return err
	}

	var savedNetworks []SavedNetwork
	json.Unmarshal(data, &savedNetworks)

	// remove the network
	found := false
	for i, network := range savedNetworks {
		if network.SSID == ssid {
			savedNetworks = slices.Delete(savedNetworks, i, i+1)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("SSID '%s' not found.", ssid)
	}

	// save back to file
	file, err := json.MarshalIndent(savedNetworks, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(savedNetworksFile, file, 0644)
}

// Turns the system's wifi on
func TurnOnWifi() error {
	cmd := exec.Command("nmcli", "radio", "wifi", "on")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to enable wifi: %w, output: %s", err, output)
	}

	return nil
}

// Turns the system's wifi off
func TurnOffWifi() error {
	cmd := exec.Command("nmcli", "radio", "wifi", "off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to enable wifi: %w, output: %s", err, output)
	}

	return nil
}
