package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/huh/spinner"
)

func showHelpMessage() {
	fmt.Println("Usage: ewc | ewc [Option]")
	fmt.Println("Options:")
	fmt.Println(" on:  turns on the wifi")
	fmt.Println(" off:  turns off the wifi")
	fmt.Println(" forget <SSID>:  forgets the provided SSID")
	fmt.Println(" help:  shows this message")
}

func waitForScan(timeoutSeconds int) {
	action := func() {
		time.Sleep(time.Duration(timeoutSeconds) * time.Second)
	}
	if err := spinner.New().Title("Scanning the nearby networks...").Action(action).Run(); err != nil {
		fmt.Println("Spinner scan failed:", err)
	}
}

func waitForConnection(timeoutSeconds int) {
	action := func() {
		time.Sleep(time.Duration(timeoutSeconds) * time.Second)
	}
	if err := spinner.New().Title("Establishing connection...").Action(action).Run(); err != nil {
		fmt.Println("Spinner establish failed:", err)
	}
}
