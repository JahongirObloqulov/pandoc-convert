package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	goRuntime "runtime"
	"strings"

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
	// 1. Dastur ijro etilayotgan papkadagi pandoc/pandoc.exe (portable/bundled rejim)
	exePath, err := os.Executable()
	if err == nil {
		localDir := filepath.Dir(exePath)
		filename := "pandoc"
		if goRuntime.GOOS == "windows" {
			filename = "pandoc.exe"
		}
		bundledPandoc := filepath.Join(localDir, filename)
		if _, err := os.Stat(bundledPandoc); err == nil {
			return bundledPandoc
		}
	}

	// 2. PATH dagi "pandoc" ni tekshiramiz
	_, err = exec.LookPath("pandoc")
	if err == nil {
		return "pandoc"
	}

	// 3. macOS/Linux uchun keng tarqalgan standart yo'llarni tekshiramiz (chunki GUI ilovalarda PATH farq qilishi mumkin)
	if goRuntime.GOOS != "windows" {
		commonPaths := []string{
			"/opt/homebrew/bin/pandoc",
			"/usr/local/bin/pandoc",
			"/usr/bin/pandoc",
			"/bin/pandoc",
		}
		for _, p := range commonPaths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	// 4. Foydalanuvchining Local AppData yo'lidagi Pandoc ni tekshiramiz (Windows uchun)
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
		tempHTMLPath := filepath.Join(filepath.Dir(outputFile), "temp_pdf_convert.html")
		tempPDFPath := filepath.Join(filepath.Dir(outputFile), "temp_pdf_convert.pdf")

		// Har qanday eski vaqtinchalik PDF faylni o'chirib tashlaymiz
		os.Remove(tempPDFPath)

		// Kirish faylini avval HTML formatiga o'giramiz (MathML va Standalone bilan)
		htmlResult := a.ConvertFile(inputFile, tempHTMLPath)
		if htmlResult == "success" {
			err := convertHTMLToPDFPlatform(tempHTMLPath, tempPDFPath)
			os.Remove(tempHTMLPath)
			if err == nil {
				// Vaqtinchalik PDF faylni yakuniy manzilga o'tkazamiz
				// Agar yakuniy fayl ochiq bo'lsa (masalan, Adobe Acrobat-da), rename xatolik beradi
				os.Remove(outputFile) // Oldingi faylni o'chiramiz
				errRename := os.Rename(tempPDFPath, outputFile)
				if errRename == nil {
					return "success"
				}
				os.Remove(tempPDFPath) // Xatolik bo'lsa tozalaymiz
				return fmt.Sprintf("PDF faylni saqlashda xatolik (ehtimol u Adobe Reader yoki boshqa dasturda ochiq turibdi): %s", errRename.Error())
			}
			os.Remove(tempPDFPath)
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

	// Chiqish fayli .txt (plain text) bo'lsa, matematik formulalarni o'qiladigan Unicode ko'rinishiga keltiramiz
	if ext == ".txt" {
		data, errRead := os.ReadFile(outputFile)
		if errRead == nil {
			cleanedContent := formatMathInText(string(data))
			_ = os.WriteFile(outputFile, []byte(cleanedContent), 0644)
		}
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
		f, err := os.Open(filePath)
		if err == nil {
			defer f.Close()
			buf := make([]byte, 4096)
			n, errRead := f.Read(buf)
			if errRead == nil || n > 0 {
				preview := string(buf[:n])
				// Agar input LaTeX (.tex) yoki Markdown (.md) bo'lsa, preview uchun ham formulalarni tozalaymiz
				if ext == ".tex" || ext == ".md" {
					preview = formatMathInText(preview)
				}
				runes := []rune(preview)
				if len(runes) > 600 {
					preview = string(runes[:600]) + "..."
				} else if n == 4096 {
					preview = string(runes) + "..."
				}
				result["preview"] = preview
			}
		}
	}

	return result
}

// cleanLaTeXMath LaTeX formulalarini odam tushunadigan oddiy matnga aylantiradi
func cleanLaTeXMath(math string) string {
	// Bo'shliqlar, quad va LaTeX formatlash yordamchilarini tozalash
	math = strings.ReplaceAll(math, `\,`, "")
	math = strings.ReplaceAll(math, `\quad`, " ")
	math = strings.ReplaceAll(math, `\qquad`, "   ")
	math = strings.ReplaceAll(math, `\dots`, "...")
	math = strings.ReplaceAll(math, `\left(`, "(")
	math = strings.ReplaceAll(math, `\right)`, ")")
	math = strings.ReplaceAll(math, `\left[`, "[")
	math = strings.ReplaceAll(math, `\right]`, "]")

	// Grek harflari va matematik simvollar
	math = strings.ReplaceAll(math, `\Gamma`, "Gamma")
	math = strings.ReplaceAll(math, `\partial`, "∂")
	math = strings.ReplaceAll(math, `\alpha`, "α")
	math = strings.ReplaceAll(math, `\beta`, "β")
	math = strings.ReplaceAll(math, `\gamma`, "γ")
	math = strings.ReplaceAll(math, `\theta`, "θ")
	math = strings.ReplaceAll(math, `\lambda`, "λ")
	math = strings.ReplaceAll(math, `\pi`, "π")
	math = strings.ReplaceAll(math, `\sigma`, "σ")
	math = strings.ReplaceAll(math, `\omega`, "ω")
	math = strings.ReplaceAll(math, `\Delta`, "Δ")
	math = strings.ReplaceAll(math, `\nabla`, "∇")
	math = strings.ReplaceAll(math, `\cdot`, " * ")
	math = strings.ReplaceAll(math, `\times`, " * ")
	math = strings.ReplaceAll(math, `\neq`, " != ")
	
	// Qo'shimcha matematik simvollar va munosabatlar
	math = strings.ReplaceAll(math, `\leq`, " <= ")
	math = strings.ReplaceAll(math, `\le`, " <= ")
	math = strings.ReplaceAll(math, `\geq`, " >= ")
	math = strings.ReplaceAll(math, `\ge`, " >= ")
	math = strings.ReplaceAll(math, `\infty`, "∞")
	math = strings.ReplaceAll(math, `\approx`, "≈")
	math = strings.ReplaceAll(math, `\pm`, "±")
	math = strings.ReplaceAll(math, `\div`, "/")
	math = strings.ReplaceAll(math, `\sum`, "Σ")
	math = strings.ReplaceAll(math, `\int`, "∫")

	// Daraja va indeks qavslarini tozalash (masalan, ^{2} -> ^2, _{u} -> _u)
	reSup := regexp.MustCompile(`\^\{([^}]+)\}`)
	math = reSup.ReplaceAllString(math, "^$1")

	reSub := regexp.MustCompile(`_\{([^}]+)\}`)
	math = reSub.ReplaceAllString(math, "_$1")

	// \frac{A}{B} kasrini [A] / [B] ko'rinishiga o'tkazish
	reFrac := regexp.MustCompile(`\\frac\{([^{}]+)\}\{([^{}]+)\}`)
	for {
		next := reFrac.ReplaceAllString(math, "[$1] / [$2]")
		if next == math {
			break
		}
		math = next
	}

	// \sqrt{A} ildizini √[A] ko'rinishiga o'tkazish
	reSqrt := regexp.MustCompile(`\\sqrt\{([^{}]+)\}`)
	for {
		next := reSqrt.ReplaceAllString(math, "√[$1]")
		if next == math {
			break
		}
		math = next
	}

	// Qolgan LaTeX qiyshiq chiziqlarini (backslash) tozalash
	math = strings.ReplaceAll(math, `\`, "")

	// Ko'p ketma-ket kelgan bo'shliqlarni bittaga keltiramiz va chetlarini tozalaymiz
	reSpaces := regexp.MustCompile(`\s+`)
	math = reSpaces.ReplaceAllString(math, " ")
	math = strings.TrimSpace(math)

	return math
}

// isMathFormula matn haqiqatan ham matematik formula ekanini tekshiradi
func isMathFormula(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	// LaTeX buyrug'i bo'lsa (backslash bo'lsa)
	if strings.Contains(s, "\\") {
		return true
	}
	// Daraja yoki indeks bo'lsa
	if strings.ContainsAny(s, "^_") {
		return true
	}
	// Matematik operatorlar va munosabatlar bo'lsa
	if strings.ContainsAny(s, "=<>+*/") {
		return true
	}
	// Minus belgisi bo'lib, yonida harflar qatnashgan bo'lsa (masalan, x-y)
	if strings.Contains(s, "-") {
		hasLetters := false
		for _, r := range s {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r > 127 {
				hasLetters = true
				break
			}
		}
		if hasLetters {
			return true
		}
	}
	// Yagona o'zgaruvchi harf bo'lsa (masalan, x, N, θ)
	runes := []rune(s)
	if len(runes) == 1 {
		r := runes[0]
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r > 127 {
			return true
		}
	}
	return false
}

// formatMathInText matn ichidagi $$...$$ va $...$ formulalarni tozalaydi
func formatMathInText(content string) string {
	// 1. Blok formulalar $$...$$
	reBlock := regexp.MustCompile(`(?s)\$\$(.*?)\$\$`)
	content = reBlock.ReplaceAllStringFunc(content, func(m string) string {
		inner := m[2 : len(m)-2]
		return cleanLaTeXMath(inner)
	})

	// 2. Qator ichidagi formulalar $...$
	// Bo'shliqlar bilan boshlanmaydigan va tugamaydigan qismlarni qidiramiz
	reInline := regexp.MustCompile(`\$([^\$\s](?:[^\$]*?[^\$\s])?)\$`)
	content = reInline.ReplaceAllStringFunc(content, func(m string) string {
		inner := m[1 : len(m)-1]
		if isMathFormula(inner) {
			return cleanLaTeXMath(inner)
		}
		return m // Matematik formula bo'lmasa, o'zgarishsiz qoldiramiz
	})

	return content
}
