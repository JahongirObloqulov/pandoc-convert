//go:build !windows
package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
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

func convertHTMLToPDFPlatform(htmlPath string, pdfPath string) error {
	browserPath := ""
	if runtime.GOOS == "darwin" {
		paths := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				browserPath = p
				break
			}
		}
	} else {
		// Linux
		browsers := []string{"google-chrome", "chromium-browser", "chromium", "google-chrome-stable"}
		for _, b := range browsers {
			if p, err := exec.LookPath(b); err == nil {
				browserPath = p
				break
			}
		}
	}

	if browserPath == "" {
		return fmt.Errorf("no headless browser found on macOS/Linux")
	}

	// Format local html file path as a file:// URL to handle spaces and special chars correctly
	u := &url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(htmlPath),
	}
	htmlURL := u.String()

	// Create a temporary directory for user data to isolate the process
	tempDir, err := os.MkdirTemp("", "pandoc-pro-pdf-*")
	if err != nil {
		// Fallback without user-data-dir if temp dir creation fails
		cmd := exec.Command(browserPath, "--headless", "--disable-gpu", "--no-sandbox", "--no-pdf-header-footer", "--print-to-pdf-no-header", "--print-to-pdf="+pdfPath, htmlURL)
		errRun := cmd.Run()
		if errRun != nil {
			return errRun
		}
	} else {
		defer os.RemoveAll(tempDir)
		cmd := exec.Command(browserPath, "--headless", "--disable-gpu", "--no-sandbox", "--no-first-run", "--no-default-browser-check", "--no-pdf-header-footer", "--print-to-pdf-no-header", "--user-data-dir="+tempDir, "--print-to-pdf="+pdfPath, htmlURL)
		errRun := cmd.Run()
		if errRun != nil {
			return errRun
		}
	}

	// Because browser launcher might exit immediately (process merging), we poll and wait for the PDF file to be created and contain data
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
