//go:build windows
package main

import (
	"os/exec"
	"path/filepath"
	"syscall"
)

// hideConsoleWindow hides the console window for a command on Windows.
func hideConsoleWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
}

func openFilePlatform(filePath string) error {
	cleanedPath := filepath.FromSlash(filePath)
	cmd := exec.Command("cmd", "/c", "start", "", cleanedPath)
	hideConsoleWindow(cmd)
	return cmd.Start()
}

func revealInExplorerPlatform(filePath string) error {
	cleanedPath := filepath.FromSlash(filePath)
	cmd := exec.Command("explorer", "/select,", cleanedPath)
	hideConsoleWindow(cmd)
	return cmd.Start()
}
