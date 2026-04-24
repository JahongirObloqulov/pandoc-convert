package main

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Fayl tanlash oynasi
func (a *App) SelectInputFile() string {
	filepath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "O'giriladigan faylni tanlang",
	})
	if err != nil {
		return ""
	}
	return filepath
}

// Saqlash joyini tanlash oynasi
func (a *App) SelectOutputFile(extension string) string {
	filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Natijani saqlash",
		DefaultFilename: "natija" + extension,
		Filters: []runtime.FileFilter{
			{DisplayName: extension + " hujjat", Pattern: "*" + extension},
			{DisplayName: "Barcha fayllar", Pattern: "*.*"},
		},
	})
	if err != nil {
		return ""
	}
	return filepath
}

// Pandoc buyrug'ini ishga tushirish
func (a *App) ConvertFile(inputFile string, outputFile string) string {
	cmd := exec.Command("pandoc", inputFile, "-o", outputFile)

	// Xatoliklarni ushlab olish
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Xatolik (kod %s): %s", err.Error(), string(output))
	}
	return "success"
}
