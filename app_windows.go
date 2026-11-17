//go:build windows

package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"
)

// hideConsoleWindow hides the console window for a command on Windows.
func hideConsoleWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
}

// openFilePlatform faylni standart dasturda ochadi.
// ShellExecute to'g'ridan-to'g'ri chaqiriladi: "cmd /c start" ishlatilmaydi, shuning uchun
// fayl nomidagi &, ^, | kabi belgilar buyruq sifatida talqin qilinmaydi.
func openFilePlatform(filePath string) error {
	cleanedPath := filepath.Clean(filepath.FromSlash(filePath))
	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	file, err := windows.UTF16PtrFromString(cleanedPath)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, verb, file, nil, nil, windows.SW_SHOWNORMAL)
}

// revealInExplorerPlatform faylni Explorer'da belgilangan holda ko'rsatadi.
func revealInExplorerPlatform(filePath string) error {
	cleanedPath := filepath.Clean(filepath.FromSlash(filePath))
	// explorer.exe argumentlarni o'ziga xos tarzda o'qiydi: /select,"yo'l" ko'rinishi kerak.
	// Windows'da fayl nomida qo'shtirnoq (") bo'lishi mumkin emas, shuning uchun bu xavfsiz.
	cmd := exec.Command("explorer.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: fmt.Sprintf(`explorer.exe /select,"%s"`, cleanedPath),
	}
	return cmd.Start()
}

// findHeadlessBrowser PDF yaratish uchun Chromium asosidagi brauzerni qidiradi.
func findHeadlessBrowser() (string, error) {
	var candidates []string
	for _, env := range []string{"ProgramFiles(x86)", "ProgramFiles", "LocalAppData"} {
		base := os.Getenv(env)
		if base == "" {
			continue
		}
		candidates = append(candidates,
			filepath.Join(base, `Microsoft\Edge\Application\msedge.exe`),
			filepath.Join(base, `Google\Chrome\Application\chrome.exe`),
			filepath.Join(base, `BraveSoftware\Brave-Browser\Application\brave.exe`),
		)
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	for _, name := range []string{"msedge", "chrome", "brave"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no headless browser found on Windows")
}

// htmlFileURL Windows yo'lini to'g'ri file:///C:/... URL'ga aylantiradi (bo'shliq va maxsus belgilar kodlanadi).
func htmlFileURL(path string) string {
	u := &url.URL{Scheme: "file", Path: "/" + filepath.ToSlash(path)}
	return u.String()
}
