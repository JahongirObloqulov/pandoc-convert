package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Headless brauzer uchun maksimal kutish vaqti
const browserTimeout = 2 * time.Minute

// convertHTMLToPDF HTML faylni o'rnatilgan Chromium asosidagi brauzer
// (Edge, Chrome, Brave, Chromium) yordamida headless rejimda PDF'ga chop etadi.
// Brauzerni topish (findHeadlessBrowser) va file:// URL yasash (htmlFileURL) platformaga xos fayllarda.
func convertHTMLToPDF(htmlPath, pdfPath string) error {
	browserPath, err := findHeadlessBrowser()
	if err != nil {
		return err
	}

	// Alohida profil papkasi: brauzer allaqachon ochiq bo'lsa ham jarayon unga
	// qo'shilib ketmaydi va cmd.Run() PDF tayyor bo'lgunicha kutadi
	profileDir, err := os.MkdirTemp("", "pandoc-pro-browser-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(profileDir)

	ctx, cancel := context.WithTimeout(context.Background(), browserTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, browserPath,
		"--headless",
		"--disable-gpu",
		"--no-first-run",
		"--no-default-browser-check",
		"--no-pdf-header-footer",
		"--print-to-pdf-no-header",
		"--user-data-dir="+profileDir,
		"--print-to-pdf="+pdfPath,
		htmlFileURL(htmlPath),
	)
	hideConsoleWindow(cmd)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("headless browser timed out after %v", browserTimeout)
		}
		return err
	}

	// Ba'zi brauzerlar faylni yozishni jarayon tugaganidan keyin yakunlaydi — qisqa kutamiz
	for i := 0; i < 50; i++ {
		if info, err := os.Stat(pdfPath); err == nil && info.Size() > 0 {
			time.Sleep(300 * time.Millisecond) // fayl qulfi bo'shashi uchun
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for PDF generation by headless browser")
}
