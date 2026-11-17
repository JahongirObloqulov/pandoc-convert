package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCleanLaTeXMath(t *testing.T) {
	cases := []struct{ in, want string }{
		{`\alpha + \beta`, "α + β"},
		{`x^{2} + y_{i}`, "x^2 + y_i"},
		{`\frac{a}{b}`, "[a] / [b]"},
		{`\frac{\frac{1}{2}}{3}`, "[[1] / [2]] / [3]"},
		{`\sqrt{x}`, "√[x]"},
		// Oldin xato chiqqan holatlar: qisqa buyruq uzunroq buyruq ichida almashtirilardi
		{`a_1 + \cdots + a_n`, "a_1 + ... + a_n"},
		{`1, 2, \ldots, n`, "1, 2, ..., n"},
		{`\left\{ x \right\}`, "{ x }"},
		{`\left( x \right)`, "( x )"},
		{`\int_0^\infty f`, "∫_0^∞ f"},
		{`x \in A`, "x ∈ A"},
		{`a \leq b \geq c`, "a <= b >= c"},
		{`a \le b`, "a <= b"},
		{`\Gamma(\gamma)`, "Γ(γ)"},
		{`a\,b`, "ab"},
		{`\text{if } x`, "if x"},
		{`\sin x`, "sin x"},
	}
	for _, c := range cases {
		if got := cleanLaTeXMath(c.in); got != c.want {
			t.Errorf("cleanLaTeXMath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatMathInText(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Formula: $x^{2}$.", "Formula: x^2."},
		{"$$\\frac{a}{b}$$", "[a] / [b]"},
		// Pul summalari formula deb tushunilmasligi kerak
		{"Narxi $5 va $10", "Narxi $5 va $10"},
		{"$5-$10", "$5-$10"},
		{"Bitta o'zgaruvchi $x$ bor", "Bitta o'zgaruvchi x bor"},
	}
	for _, c := range cases {
		if got := formatMathInText(c.in); got != c.want {
			t.Errorf("formatMathInText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSamePath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if !samePath(file, file) {
		t.Error("bir xil yo'l bir xil deb topilishi kerak")
	}
	if !samePath(file, filepath.Join(dir, ".", "doc.md")) {
		t.Error("normallashtirilgan yo'l bir xil deb topilishi kerak")
	}
	if samePath(file, filepath.Join(dir, "doc.docx")) {
		t.Error("turli fayllar farqli bo'lishi kerak")
	}
	if runtime.GOOS == "windows" && !samePath(file, strings.ToUpper(file)) {
		t.Error("Windows'da katta-kichik harf farqi e'tiborga olinmasligi kerak")
	}
}

func TestConvertFileRejectsSameFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "doc.md")
	original := []byte("# Asl matn\n")
	if err := os.WriteFile(file, original, 0644); err != nil {
		t.Fatal(err)
	}

	if got := NewApp().ConvertFile(file, file); got != ErrSameFile {
		t.Fatalf("ConvertFile(same, same) = %q, want %q", got, ErrSameFile)
	}
	data, _ := os.ReadFile(file)
	if string(data) != string(original) {
		t.Error("asl fayl o'zgarmasligi kerak")
	}
}

func TestReadPreviewTrimsBrokenUTF8(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "long.txt")
	// "ў" 2 bayt: 4095 ta "a" + "ў" => bufer chegarasi belgining o'rtasiga tushadi
	content := strings.Repeat("a", previewReadBytes-1) + "ўўў"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	preview, ok := readPreview(file, ".txt")
	if !ok {
		t.Fatal("preview olinmadi")
	}
	if strings.ContainsRune(preview, '\uFFFD') {
		t.Error("preview ichida buzilgan UTF-8 belgi (�) bo'lmasligi kerak")
	}
}

func TestGetFileInfoUppercaseExtension(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "README.MD")
	if err := os.WriteFile(file, []byte("# Salom"), 0644); err != nil {
		t.Fatal(err)
	}
	info := NewApp().GetFileInfo(file)
	if info["extension"] != ".md" {
		t.Errorf("extension = %v, want .md", info["extension"])
	}
	if _, ok := info["preview"]; !ok {
		t.Error(".MD fayl uchun ham preview bo'lishi kerak")
	}
}
