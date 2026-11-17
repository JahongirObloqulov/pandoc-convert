//go:build !windows

package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// hideConsoleWindow is a no-op on non-Windows systems since they do not experience flashing console windows.
func hideConsoleWindow(cmd *exec.Cmd) {}

func openFilePlatform(filePath string) error {
	cleanedPath := filepath.Clean(filePath)
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", cleanedPath)
	} else {
		cmd = exec.Command("xdg-open", cleanedPath)
	}
	return cmd.Start()
}

func revealInExplorerPlatform(filePath string) error {
	cleanedPath := filepath.Clean(filePath)
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", "-R", cleanedPath)
	} else {
		// Linux: Open parent folder and rely on the file manager, as there is no uniform file highlighting standard
		cmd = exec.Command("xdg-open", filepath.Dir(cleanedPath))
	}
	return cmd.Start()
}

// findHeadlessBrowser PDF yaratish uchun Chromium asosidagi brauzerni qidiradi.
func findHeadlessBrowser() (string, error) {
	if runtime.GOOS == "darwin" {
		paths := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
	} else {
		browsers := []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "microsoft-edge", "brave-browser"}
		for _, b := range browsers {
			if p, err := exec.LookPath(b); err == nil {
				return p, nil
			}
		}
	}
	return "", fmt.Errorf("no headless browser found on macOS/Linux")
}

// htmlFileURL yo'lni file:// URL'ga aylantiradi (bo'shliq va maxsus belgilar kodlanadi).
func htmlFileURL(path string) string {
	u := &url.URL{Scheme: "file", Path: filepath.ToSlash(path)}
	return u.String()
}
