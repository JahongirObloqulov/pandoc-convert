//go:build !windows
package main

import (
	"os/exec"
	"path/filepath"
	"runtime"
)

// hideConsoleWindow is a no-op on non-Windows systems since they do not experience flashing console windows.
func hideConsoleWindow(cmd *exec.Cmd) {
}

func openFilePlatform(filePath string) error {
	cleanedPath := filepath.FromSlash(filePath)
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", cleanedPath)
	} else {
		cmd = exec.Command("xdg-open", cleanedPath)
	}
	return cmd.Start()
}

func revealInExplorerPlatform(filePath string) error {
	cleanedPath := filepath.FromSlash(filePath)
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", "-R", cleanedPath)
	} else {
		// Linux: Open parent folder and rely on the file manager, as there is no uniform file highlighting standard
		parentDir := filepath.Dir(cleanedPath)
		cmd = exec.Command("xdg-open", parentDir)
	}
	return cmd.Start()
}
