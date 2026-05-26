//go:build windows
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
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
		errRun := cmd.Run()
		if errRun != nil {
			return errRun
		}
	} else {
		defer os.RemoveAll(tempDir)
		cmd := exec.Command(edgePath, "--headless", "--disable-gpu", "--no-sandbox", "--no-first-run", "--no-default-browser-check", "--user-data-dir="+tempDir, "--print-to-pdf="+pdfPath, htmlPath)
		hideConsoleWindow(cmd)
		errRun := cmd.Run()
		if errRun != nil {
			return errRun
		}
	}

	// Because Edge launcher might exit immediately (process merging), we poll and wait for the PDF file to be created and contain data
	for i := 0; i < 50; i++ {
		if info, err := os.Stat(pdfPath); err == nil && info.Size() > 0 {
			// Give it a tiny extra time to release file locks
			time.Sleep(300 * time.Millisecond)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for PDF generation by headless browser")
}
