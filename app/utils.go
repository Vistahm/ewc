package app

import (
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/godbus/dbus/v5"
)

// Prints out the help message for the help argument
func ShowHelpMessage() {
	fmt.Println("Usage: ewc | ewc [Option]")
	fmt.Println("Options:")
	fmt.Println(" on:  turns on the wifi")
	fmt.Println(" off:  turns off the wifi")
	fmt.Println(" forget <SSID>:  forgets the provided SSID")
	fmt.Println(" help:  shows this message")
}

// Handles the error as a helper function
func HandleError(err error, message string) {
	if err != nil {
		fmt.Printf("%s: %s\n", err, message)
		os.Exit(1)
	}
}

// Waits for scanning the nerby networks based on the given timeoutSeconds
func WaitForScan(timeoutSeconds int) {
	action := func() {
		time.Sleep(time.Duration(timeoutSeconds) * time.Second)
	}
	if err := spinner.New().Title("Scanning the nearby networks...").Action(action).Run(); err != nil {
		fmt.Println("Spinner scan failed:", err)
	}
}

// Waits to establish connection based on the given timeoutSeconds
func WaitForConnection(timeoutSeconds int) {
	action := func() {
		time.Sleep(time.Duration(timeoutSeconds) * time.Second)
	}
	if err := spinner.New().Title("Establishing connection...").Action(action).Run(); err != nil {
		fmt.Println("Spinner establish failed:", err)
	}
}

// Handles the command-line arguments
func HandleArguments(args []string) {

	if !slices.Equal(args, nil) {
		switch args[0] {
		case "on":
			if err := TurnOnWifi(); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Wi-Fi Enabled.")
			}
			os.Exit(0)

		case "off":
			if err := TurnOffWifi(); err != nil {
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
			if err := ForgetNetwork(ssidToForget); err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("Success.\nForgotten network: %s\n", ssidToForget)
			}
			os.Exit(0)

		case "help":
			ShowHelpMessage()
			os.Exit(0)

		default:
			fmt.Println("Unknown command. Use 'help' for a list of commands.")
			os.Exit(0)
		}
	}
}

// Checks the system's NetworkManager state
func GetNetworkManagerState(obj dbus.BusObject) error {
	var state uint32
	err := obj.Call("org.freedesktop.NetworkManager.state", 0).Store(&state)
	if err != nil {
		return err
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

	return nil
}

// Creates a huh form to accept an access point from user
func SelectAccessPoint(accessPoints []AccessPoint) AccessPoint {
	var selectedAP AccessPoint
	var ssidOptions []huh.Option[AccessPoint]

	for _, ap := range accessPoints {
		ssidOptions = append(ssidOptions, huh.NewOption(fmt.Sprintf("%s (Strength: %d)", ap.SSID, ap.Strength), ap))
	}

	selectForm := huh.NewSelect[AccessPoint]().
		Title("Select Wi-Fi Network").
		Options(ssidOptions...).
		Value(&selectedAP)

	HandleError(selectForm.Run(), "Error with form")
	return selectedAP
}

// If a saved password was found for the selectedAP it will load it, otherwise it will prompt the user for a password. Also if the the access point is not protected with WPA/WPA2, ignore the password prompt.
func GetPasswordForAccessPoint(selectedAP AccessPoint) string {
	var password string
	savedPassword, found := LoadPassword(selectedAP.SSID)

	if (selectedAP.Flags & 0x1) > 0 {
		if found {
			password = savedPassword
			fmt.Println("Using saved password for:", selectedAP.SSID)
		} else {
			password = PromptForPassword()
		}
	} else {
		fmt.Println("No poassword required for this network.")
	}

	return password
}

// Creates a huh form to accept password for the selected ssid
func PromptForPassword() string {
	var passwordInput string
	var showPassword bool
	var passwordForm *huh.Input

	showPasswordToggle := huh.NewConfirm().
		Title("Show Password?").
		Value(&showPassword).
		Run()
	if showPasswordToggle != nil {
		fmt.Println("failed to showPasswordToggle")
		os.Exit(1)
	}

	if showPassword {
		passwordForm = huh.NewInput().Title("Enter Password:").EchoMode(huh.EchoModeNormal).Value(&passwordInput)
	} else {
		passwordForm = huh.NewInput().Title("Enter Password:").EchoMode(huh.EchoModePassword).Value(&passwordInput)
	}

	HandleError(passwordForm.Run(), "failed to load form")

	return passwordInput
}
