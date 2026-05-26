//go:build !windows
package main

import (
	"fmt"
	"os"
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

func convertHTMLToPDFPlatform(htmlPath string, pdfPath string) error {
	browserPath := ""
	if runtime.GOOS == "darwin" {
		paths := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
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

	cmd := exec.Command(browserPath, "--headless", "--disable-gpu", "--no-sandbox", "--print-to-pdf="+pdfPath, htmlPath)
	return cmd.Run()
}
