//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// init provides the normal Windows double-click experience while preserving
// the existing CLI when any arguments are supplied.
func init() {
	if len(os.Args) > 1 {
		return
	}

	exe, err := os.Executable()
	if err == nil {
		_ = os.Chdir(filepath.Dir(exe))
	}

	const configPath = "automation.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := defaultConfig()
		if saveErr := saveConfig(configPath, cfg); saveErr != nil {
			showStartupError(fmt.Errorf("create starter config: %w", saveErr))
			os.Exit(1)
		}
	}

	go func() {
		time.Sleep(700 * time.Millisecond)
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://127.0.0.1:17654").Start()
	}()

	fmt.Println("IdleOn Account Agent")
	fmt.Println("Launching dashboard: http://127.0.0.1:17654")
	fmt.Println("Keep this window open while using the agent. Press Ctrl+C to close it.")
	if err := cmdServe([]string{"-config", configPath, "-snapshot", "auto", "-port", "17654"}); err != nil {
		showStartupError(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func showStartupError(err error) {
	message := "IdleOn Account Agent could not start:\n\n" + err.Error() + "\n\nOpen Command Prompt in this folder and run:\nIdleOn-Account-Agent.exe doctor -config automation.json"
	_ = exec.Command("powershell.exe", "-NoProfile", "-Command", "Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show($args[0], 'IdleOn Account Agent')", message).Run()
	fmt.Fprintln(os.Stderr, message)
}
