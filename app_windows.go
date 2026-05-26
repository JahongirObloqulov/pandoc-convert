//go:build windows
package main

import (
	"fmt"
	"os"
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

func convertHTMLToPDFPlatform(htmlPath string, pdfPath string) error {
	edgePath := ""
	paths := []string{
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			edgePath = p
			break
		}
	}

	if edgePath == "" {
		if p, err := exec.LookPath("msedge"); err == nil {
			edgePath = p
		} else if p, err := exec.LookPath("chrome"); err == nil {
			edgePath = p
		} else {
			return fmt.Errorf("no headless browser found on Windows")
		}
	}

	// Create a temporary directory for user data to isolate the process
	tempDir, err := os.MkdirTemp("", "pandoc-pro-pdf-*")
	if err != nil {
		// Fallback without user-data-dir if temp dir creation fails
		cmd := exec.Command(edgePath, "--headless", "--disable-gpu", "--no-sandbox", "--print-to-pdf="+pdfPath, htmlPath)
		hideConsoleWindow(cmd)
		return cmd.Run()
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.Command(edgePath, "--headless", "--disable-gpu", "--no-sandbox", "--no-first-run", "--no-default-browser-check", "--user-data-dir="+tempDir, "--print-to-pdf="+pdfPath, htmlPath)
	hideConsoleWindow(cmd)
	return cmd.Run()
}
