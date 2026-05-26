package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
		Filters: []runtime.FileFilter{
			{DisplayName: "Hujjatlar (*.md, *.txt, *.html, *.docx, *.epub, *.tex)", Pattern: "*.md;*.txt;*.html;*.docx;*.epub;*.tex"},
			{DisplayName: "Barcha fayllar", Pattern: "*.*"},
		},
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

// Pandoc buyrug'ining yo'lini aniqlash
func (a *App) getPandocPath() string {
	// 1. Dastur ijro etilayotgan papkadagi pandoc.exe (portable/bundled rejim)
	exePath, err := os.Executable()
	if err == nil {
		localDir := filepath.Dir(exePath)
		bundledPandoc := filepath.Join(localDir, "pandoc.exe")
		if _, err := os.Stat(bundledPandoc); err == nil {
			return bundledPandoc
		}
	}

	// 2. PATH dagi "pandoc" ni tekshiramiz
	_, err = exec.LookPath("pandoc")
	if err == nil {
		return "pandoc"
	}

	// 3. Foydalanuvchining Local AppData yo'lidagi Pandoc ni tekshiramiz
	home, err := os.UserHomeDir()
	if err == nil {
		localPandoc := filepath.Join(home, "AppData", "Local", "Pandoc", "pandoc.exe")
		if _, err := os.Stat(localPandoc); err == nil {
			return localPandoc
		}
	}

	// Oxirgi chora: odatiy yo'l
	return "pandoc"
}

// Pandoc o'rnatilganini tekshirish
func (a *App) CheckPandocInstalled() bool {
	path := a.getPandocPath()
	cmd := exec.Command(path, "--version")
	hideConsoleWindow(cmd)
	err := cmd.Run()
	return err == nil
}

// Pandoc buyrug'ini ishga tushirish
func (a *App) ConvertFile(inputFile string, outputFile string) string {
	path := a.getPandocPath()
	ext := filepath.Ext(outputFile)

	// PDF konvertatsiya uchun oflayn brauzer orqali HTML-to-PDF muqobil yo'lini sinab ko'ramiz
	if ext == ".pdf" {
		tempHTMLPath := filepath.Join(filepath.Dir(outputFile), "~temp_pdf_convert.html")

		// Kirish faylini avval HTML formatiga o'giramiz (MathML va Standalone bilan)
		htmlResult := a.ConvertFile(inputFile, tempHTMLPath)
		if htmlResult == "success" {
			err := convertHTMLToPDFPlatform(tempHTMLPath, outputFile)
			os.Remove(tempHTMLPath)
			if err == nil {
				return "success"
			}
			// Muqobil yo'l xato bersa, standart Pandoc/LaTeX yo'liga o'tiladi (fallback)
		}
	}

	args := []string{inputFile, "-o", outputFile}
	if ext == ".html" || ext == ".htm" {
		// Standalone HTML fayl yaratish va formulalarni MathML formatida oflayn va chiroyli ko'rsatish
		args = append(args, "-s", "--mathml")
	}

	cmd := exec.Command(path, args...)
	hideConsoleWindow(cmd)

	// Xatoliklarni ushlab olish
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Xatolik (kod %s): %s", err.Error(), string(output))
	}
	return "success"
}

// O'girilgan faylni standart dasturda ochish
func (a *App) OpenFile(filePath string) string {
	err := openFilePlatform(filePath)
	if err != nil {
		return err.Error()
	}
	return "success"
}

// Faylni Explorer papkasida belgilab ko'rsatish
func (a *App) RevealInExplorer(filePath string) string {
	err := revealInExplorerPlatform(filePath)
	if err != nil {
		return err.Error()
	}
	return "success"
}

// Fayl haqida ma'lumot olish
func (a *App) GetFileInfo(filePath string) map[string]interface{} {
	result := make(map[string]interface{})
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		result["error"] = err.Error()
		return result
	}

	result["name"] = fileInfo.Name()
	result["size"] = fileInfo.Size() // in bytes
	result["modTime"] = fileInfo.ModTime().Format("02.01.2006 15:04:05")
	result["extension"] = filepath.Ext(filePath)

	// Fayl matnli bo'lsa uni o'qish va preview olish
	ext := filepath.Ext(filePath)
	if ext == ".md" || ext == ".txt" || ext == ".html" || ext == ".tex" {
		data, err := os.ReadFile(filePath)
		if err == nil {
			preview := string(data)
			if len(preview) > 600 {
				preview = preview[:600] + "..."
			}
			result["preview"] = preview
		}
	}

	return result
}
