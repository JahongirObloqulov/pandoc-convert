package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	goRuntime "runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	// Pandoc jarayoni uchun maksimal vaqt (osilib qolishning oldini olish)
	pandocTimeout = 5 * time.Minute

	// ErrSameFile — chiqish fayli asl fayl bilan bir xil bo'lganda frontendga qaytariladigan kod
	ErrSameFile = "ERR_SAME_FILE"
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
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "O'giriladigan faylni tanlang",
		Filters: []runtime.FileFilter{
			{DisplayName: "Hujjatlar (*.md, *.txt, *.html, *.docx, *.epub, *.tex)", Pattern: "*.md;*.txt;*.html;*.docx;*.epub;*.tex"},
			{DisplayName: "Barcha fayllar", Pattern: "*.*"},
		},
	})
	if err != nil {
		return ""
	}
	return path
}

// Saqlash joyini tanlash oynasi
func (a *App) SelectOutputFile(extension string) string {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
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
	return path
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
	if _, err := exec.LookPath("pandoc"); err == nil {
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
	} else {
		// 4. Foydalanuvchining Local AppData yo'lidagi Pandoc ni tekshiramiz (Windows uchun)
		if home, err := os.UserHomeDir(); err == nil {
			localPandoc := filepath.Join(home, "AppData", "Local", "Pandoc", "pandoc.exe")
			if _, err := os.Stat(localPandoc); err == nil {
				return localPandoc
			}
		}
	}

	// Oxirgi chora: odatiy yo'l
	return "pandoc"
}

// Pandoc o'rnatilganini tekshirish
func (a *App) CheckPandocInstalled() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, a.getPandocPath(), "--version")
	hideConsoleWindow(cmd)
	return cmd.Run() == nil
}

// runPandoc Pandoc'ni timeout bilan ishga tushiradi. Nisbiy rasmlar va boshqa
// resurslar topilishi uchun ishchi papka va --resource-path asl fayl papkasiga o'rnatiladi.
func (a *App) runPandoc(inputFile, outputFile string, extraArgs ...string) error {
	inputDir := filepath.Dir(inputFile)
	args := []string{inputFile, "-o", outputFile, "--resource-path=" + inputDir}
	args = append(args, extraArgs...)

	ctx, cancel := context.WithTimeout(context.Background(), pandocTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, a.getPandocPath(), args...)
	cmd.Dir = inputDir
	hideConsoleWindow(cmd)

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("Xatolik: Pandoc %v ichida javob bermadi va to'xtatildi", pandocTimeout)
	}
	if err != nil {
		return fmt.Errorf("Xatolik (kod %s): %s", err.Error(), string(output))
	}
	return nil
}

// Pandoc buyrug'ini ishga tushirish
func (a *App) ConvertFile(inputFile string, outputFile string) string {
	if strings.TrimSpace(inputFile) == "" || strings.TrimSpace(outputFile) == "" {
		return "Xatolik: kirish yoki chiqish fayli ko'rsatilmagan"
	}

	// Absolyut yo'llar: ishchi papka o'zgarganda ham to'g'ri ishlaydi va
	// fayl nomi "-" bilan boshlanib, Pandoc opsiyasi deb tushunilishining oldini oladi
	var err error
	if inputFile, err = filepath.Abs(inputFile); err != nil {
		return "Xatolik: " + err.Error()
	}
	if outputFile, err = filepath.Abs(outputFile); err != nil {
		return "Xatolik: " + err.Error()
	}

	// Asl faylni tasodifan ustidan yozib yubormaslik
	if samePath(inputFile, outputFile) {
		return ErrSameFile
	}

	ext := strings.ToLower(filepath.Ext(outputFile))

	// PDF konvertatsiya uchun oflayn brauzer orqali HTML-to-PDF yo'lini sinab ko'ramiz
	if ext == ".pdf" {
		err := a.convertToPDFViaBrowser(inputFile, outputFile)
		if err == nil {
			return "success"
		}
		var saveErr *outputSaveError
		if errors.As(err, &saveErr) {
			return fmt.Sprintf("PDF faylni saqlashda xatolik (ehtimol u Adobe Reader yoki boshqa dasturda ochiq turibdi): %s", saveErr.err.Error())
		}
		// Muqobil yo'l xato bersa, standart Pandoc/LaTeX yo'liga o'tiladi (fallback)
	}

	var extraArgs []string
	if ext == ".html" || ext == ".htm" {
		// Standalone HTML fayl yaratish va formulalarni MathML formatida oflayn va chiroyli ko'rsatish
		extraArgs = append(extraArgs, "-s", "--mathml")
	}

	if err := a.runPandoc(inputFile, outputFile, extraArgs...); err != nil {
		return err.Error()
	}

	// Chiqish fayli .txt (plain text) bo'lsa, matematik formulalarni o'qiladigan Unicode ko'rinishiga keltiramiz
	if ext == ".txt" {
		if data, errRead := os.ReadFile(outputFile); errRead == nil {
			_ = os.WriteFile(outputFile, []byte(formatMathInText(string(data))), 0644)
		}
	}

	return "success"
}

// outputSaveError — PDF tayyor, lekin yakuniy manzilga saqlab bo'lmadi
type outputSaveError struct{ err error }

func (e *outputSaveError) Error() string { return e.err.Error() }

// convertToPDFViaBrowser hujjatni avval resurslari ichiga joylangan HTML'ga, so'ng
// headless brauzer yordamida PDF'ga o'giradi. Barcha oraliq fayllar alohida vaqtinchalik papkada yaratiladi.
func (a *App) convertToPDFViaBrowser(inputFile, outputFile string) error {
	tempDir, err := os.MkdirTemp("", "pandoc-pro-pdf-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	htmlPath := filepath.Join(tempDir, "document.html")
	pdfPath := filepath.Join(tempDir, "document.pdf")

	// --embed-resources: rasmlar HTML ichiga joylanadi, shuning uchun HTML boshqa papkada bo'lsa ham ko'rinadi
	if err := a.runPandoc(inputFile, htmlPath, "-s", "--mathml", "--embed-resources"); err != nil {
		return err
	}
	if err := convertHTMLToPDF(htmlPath, pdfPath); err != nil {
		return err
	}
	if err := moveFile(pdfPath, outputFile); err != nil {
		return &outputSaveError{err: err}
	}
	return nil
}

// moveFile faylni ko'chiradi. Turli disklar orasida (os.Rename ishlamaganda) nusxalab o'tkazadi.
func moveFile(src, dst string) error {
	// Oldingi faylni o'chiramiz. Agar u boshqa dasturda ochiq bo'lsa, Windows bu yerda xato beradi
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dst)
		return err
	}
	return out.Close()
}

// samePath ikki yo'l bitta faylga ishora qilishini tekshiradi
func samePath(a, b string) bool {
	pa, errA := filepath.Abs(a)
	pb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return false
	}
	pa, pb = filepath.Clean(pa), filepath.Clean(pb)

	// Ikkala fayl ham mavjud bo'lsa — eng ishonchli tekshiruv (symlink va h.k.)
	if sa, err := os.Stat(pa); err == nil {
		if sb, err := os.Stat(pb); err == nil {
			return os.SameFile(sa, sb)
		}
	}

	// Windows va macOS fayl tizimlari odatda katta-kichik harfni farqlamaydi
	if goRuntime.GOOS == "windows" || goRuntime.GOOS == "darwin" {
		return strings.EqualFold(pa, pb)
	}
	return pa == pb
}

// O'girilgan faylni standart dasturda ochish
func (a *App) OpenFile(filePath string) string {
	if err := openFilePlatform(filePath); err != nil {
		return err.Error()
	}
	return "success"
}

// Faylni Explorer papkasida belgilab ko'rsatish
func (a *App) RevealInExplorer(filePath string) string {
	if err := revealInExplorerPlatform(filePath); err != nil {
		return err.Error()
	}
	return "success"
}

// Preview uchun o'qiladigan maksimal bayt va ko'rsatiladigan belgilar soni
const (
	previewReadBytes = 4096
	previewMaxRunes  = 600
)

// Fayl haqida ma'lumot olish
func (a *App) GetFileInfo(filePath string) map[string]interface{} {
	result := make(map[string]interface{})
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		result["error"] = err.Error()
		return result
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	result["name"] = fileInfo.Name()
	result["size"] = fileInfo.Size() // in bytes
	result["modTime"] = fileInfo.ModTime().Format("02.01.2006 15:04:05")
	result["extension"] = ext

	// Fayl matnli bo'lsa uni o'qish va preview olish
	if ext == ".md" || ext == ".txt" || ext == ".html" || ext == ".htm" || ext == ".tex" {
		if preview, ok := readPreview(filePath, ext); ok {
			result["preview"] = preview
		}
	}

	return result
}

func readPreview(filePath, ext string) (string, bool) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", false
	}
	defer f.Close()

	buf := make([]byte, previewReadBytes)
	n, _ := io.ReadFull(f, buf)
	if n == 0 {
		return "", false
	}
	chunk := buf[:n]
	truncated := n == previewReadBytes

	// Bufer chegarasida kesilib qolgan ko'p baytli UTF-8 belgini olib tashlaymiz ("�" chiqmasligi uchun)
	for i := 0; truncated && i < utf8.UTFMax-1 && len(chunk) > 0 && !utf8.Valid(chunk); i++ {
		chunk = chunk[:len(chunk)-1]
	}

	preview := string(chunk)
	// Agar input LaTeX (.tex) yoki Markdown (.md) bo'lsa, preview uchun ham formulalarni tozalaymiz
	if ext == ".tex" || ext == ".md" {
		preview = formatMathInText(preview)
	}
	runes := []rune(preview)
	if len(runes) > previewMaxRunes {
		preview = string(runes[:previewMaxRunes]) + "..."
	} else if truncated {
		preview += "..."
	}
	return preview, true
}

// LaTeX buyruqlarining oddiy matndagi ko'rinishi.
// Buyruq nomi to'liq (so'z chegarasi bilan) solishtiriladi, shuning uchun
// \cdots, \ldots, \left, \infty kabi buyruqlar qisqaroq buyruqlar bilan aralashib ketmaydi.
var latexSymbols = map[string]string{
	// Bo'shliqlar va formatlash yordamchilari
	"quad": " ", "qquad": "   ", "left": "", "right": "",
	"dots": "...", "ldots": "...", "cdots": "...", "vdots": "...",

	// Grek harflari
	"alpha": "α", "beta": "β", "gamma": "γ", "delta": "δ", "epsilon": "ε", "varepsilon": "ε",
	"theta": "θ", "lambda": "λ", "mu": "μ", "pi": "π", "rho": "ρ", "sigma": "σ", "tau": "τ",
	"phi": "φ", "varphi": "φ", "omega": "ω",
	"Gamma": "Γ", "Delta": "Δ", "Theta": "Θ", "Lambda": "Λ", "Pi": "Π", "Sigma": "Σ", "Phi": "Φ", "Omega": "Ω",

	// Matematik simvollar va munosabatlar
	"partial": "∂", "nabla": "∇", "infty": "∞",
	"cdot": " * ", "times": " * ", "div": "/", "pm": "±", "mp": "∓",
	"neq": " != ", "ne": " != ", "leq": " <= ", "le": " <= ", "geq": " >= ", "ge": " >= ",
	"approx": "≈", "equiv": "≡", "sim": "~",
	"sum": "Σ", "prod": "Π", "int": "∫", "oint": "∮",
	"to": " → ", "rightarrow": " → ", "leftarrow": " ← ", "Rightarrow": " ⇒ ", "iff": " ⇔ ",
	"in": " ∈ ", "notin": " ∉ ", "subset": " ⊂ ", "cup": " ∪ ", "cap": " ∩ ", "forall": "∀", "exists": "∃",
}

var (
	reTextCmd    = regexp.MustCompile(`\\(?:text|mathrm|mathbf|mathit|mathsf|mathtt|operatorname)\{([^{}]*)\}`)
	reThinSpace  = regexp.MustCompile(`\\[,;:!]`)
	reCmdSpace   = regexp.MustCompile(`\\ `)
	reLatexCmd   = regexp.MustCompile(`\\([A-Za-z]+)`)
	reSup        = regexp.MustCompile(`\^\{([^}]+)\}`)
	reSub        = regexp.MustCompile(`_\{([^}]+)\}`)
	reFrac       = regexp.MustCompile(`\\frac\{([^{}]+)\}\{([^{}]+)\}`)
	reSqrt       = regexp.MustCompile(`\\sqrt\{([^{}]+)\}`)
	reSpaces     = regexp.MustCompile(`\s+`)
	reMathBlock  = regexp.MustCompile(`(?s)\$\$(.*?)\$\$`)
	reMathInline = regexp.MustCompile(`\$([^\$\s](?:[^\$]*?[^\$\s])?)\$`)
)

// replaceUntilStable qayta-qayta almashtiradi (ichma-ich joylashgan \frac, \sqrt uchun)
func replaceUntilStable(re *regexp.Regexp, s, repl string) string {
	for {
		next := re.ReplaceAllString(s, repl)
		if next == s {
			return s
		}
		s = next
	}
}

// cleanLaTeXMath LaTeX formulalarini odam tushunadigan oddiy matnga aylantiradi
func cleanLaTeXMath(math string) string {
	// \text{...}, \mathrm{...} kabi buyruqlardan faqat matnni qoldiramiz
	math = replaceUntilStable(reTextCmd, math, "$1")

	// Kichik bo'shliqlar (\, \; \: \!) va "\ " ni tozalash
	math = reThinSpace.ReplaceAllString(math, "")
	math = reCmdSpace.ReplaceAllString(math, " ")

	// Ma'lum buyruqlarni simvollarga almashtiramiz; noma'lumlari (\frac, \sqrt, \sin ...) o'zgarishsiz qoladi
	math = reLatexCmd.ReplaceAllStringFunc(math, func(m string) string {
		if repl, ok := latexSymbols[m[1:]]; ok {
			return repl
		}
		return m
	})

	// Daraja va indeks qavslarini tozalash (masalan, ^{2} -> ^2, _{u} -> _u)
	math = reSup.ReplaceAllString(math, "^$1")
	math = reSub.ReplaceAllString(math, "_$1")

	// \frac{A}{B} kasrini [A] / [B] ko'rinishiga o'tkazish
	math = replaceUntilStable(reFrac, math, "[$1] / [$2]")

	// \sqrt{A} ildizini √[A] ko'rinishiga o'tkazish
	math = replaceUntilStable(reSqrt, math, "√[$1]")

	// Qolgan LaTeX qiyshiq chiziqlarini (backslash) tozalash
	math = strings.ReplaceAll(math, `\`, "")

	// Ko'p ketma-ket kelgan bo'shliqlarni bittaga keltiramiz va chetlarini tozalaymiz
	math = reSpaces.ReplaceAllString(math, " ")
	return strings.TrimSpace(math)
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
		for _, r := range s {
			if isLetter(r) {
				return true
			}
		}
	}
	// Yagona o'zgaruvchi harf bo'lsa (masalan, x, N, θ)
	runes := []rune(s)
	return len(runes) == 1 && isLetter(runes[0])
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r > 127
}

// formatMathInText matn ichidagi $$...$$ va $...$ formulalarni tozalaydi
func formatMathInText(content string) string {
	// 1. Blok formulalar $$...$$
	content = reMathBlock.ReplaceAllStringFunc(content, func(m string) string {
		return cleanLaTeXMath(m[2 : len(m)-2])
	})

	// 2. Qator ichidagi formulalar $...$ (bo'shliq bilan boshlanmaydigan va tugamaydigan)
	content = reMathInline.ReplaceAllStringFunc(content, func(m string) string {
		inner := m[1 : len(m)-1]
		if isMathFormula(inner) {
			return cleanLaTeXMath(inner)
		}
		return m // Matematik formula bo'lmasa, o'zgarishsiz qoldiramiz
	})

	return content
}
