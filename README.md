# Pandoc Pro Konverteri

Pandoc Pro — bu Markdown, Microsoft Word (DOCX), PDF, HTML, EPUB, LaTeX va oddiy matnli (TXT) hujjatlarni o'zaro bir formatdan ikkinchisiga oson va professional darajada o'giruvchi cross-platform desktop ilova. 

Ushbu ilova **Wails (Go + HTML/JS/CSS)** texnologiyasi asosida yasalgan bo'lib, o'z ichida oflayn ishlash uchun moslashtirilgan milliy va xalqaro formatlarni qo'llab-quvvatlaydi.

---

## 🚀 Asosiy Imkoniyatlar
* **Premium Glassmorphic Interfeys**: Zamonaviy dizayn, dark mode va silliq animatsiyalar.
* **Drag & Drop**: Fayllarni interfeysga shunchaki sudrab tashlash orqali yuklash.
* **Fayl Preview**: Matnli hujjatlar tarkibini o'girishdan oldin dastur ichida tezkor ko'rish.
* **Avtomatik Taklif (Auto-suggest)**: Kengaytma tanlanganda saqlash joyi va fayl nomini avtomatik to'ldirish.
* **Tezkor Harakatlar**: Fayl o'girilgandan so'ng uni to'g'ridan-to'g'ri tizim orqali ochish yoki u joylashgan papkani Explorer/Finder orqali ko'rish.
* **Tarix Paneli**: Oxirgi 5 ta o'girilgan fayllarni tezkor kuzatish va ochish.
* **100% Oflayn Ishlash**: Dastur ishlashi uchun kompyuterda internet bo'lishi yoki tizimga Pandoc dasturini alohida o'rnatish talab etilmaydi (chunki oflayn paket ichida kerakli barcha vositalar jamlangan).

---

## 📦 Oflayn Dasturni O'rnatish va Ishga Tushirish

### 1. Windows (PC)
1. Chiqarilgan fayllar ichidan [pandoc-pro-installer.exe](build/bin/pandoc-pro-installer.exe) o'rnatuvchi faylini ishga tushiring.
2. Dastur avtomatik tarzda `C:\Program Files\PandocConvertPro` papkasiga o'rnatiladi va ish stolida yorliq hosil qiladi.
3. Dasturni ishga tushiring va oflayn rejimda ishlatishingiz mumkin.

### 2. macOS (MacBook Intel & M1/M2/M3)
macOS xavfsizlik qoidalari tufayli unsigned (imzolanmagan) dasturlarni to'g'ridan-to'g'ri ochishni bloklashi mumkin. Quyidagi bosqichlarni bajaring:
1. `pandoc-pro-mac.zip` faylini oching. Ichidan `pandoc-app.app` dasturi chiqadi.
2. Dasturni `Applications` (Dasturlar) papkasiga ko'chiring.
3. Terminal dasturini oching va quyidagi buyruqni kiriting (tizim ruxsat berishi uchun):
   ```bash
   xattr -cr /Applications/pandoc-app.app
   ```
4. Dasturni odatiy tarzda oching. Ilova ichiga macOS uchun maxsus moslashtirilgan `pandoc` fayli joylashtirilgani sababli, u internetsiz to'liq ishlaydi.

### 3. Linux (Ubuntu / Debian / Fedora va boshqalar)
1. `pandoc-pro-linux.tar.gz` faylini istalgan papkaga arxivdan chiqaring.
2. Arxiv ichidan `pandoc-app` va `pandoc` fayllari chiqadi (ular bitta papkada turishi shart).
3. Terminal orqali dasturni ishga tushirish uchun ruxsat bering:
   ```bash
   chmod +x pandoc-app pandoc
   ```
4. Dasturni ishga tushiring:
   ```bash
   ./pandoc-app
   ```

---

## 🛠️ Loyihani Mahalliy Yig'ish (Local Compilation)

Agar dasturni o'zingiz noldan yig'moqchi bo'lsangiz, kompyuteringizda **Go (1.23+)** va **Node.js (18+)** o'rnatilgan bo'lishi talab etiladi.

1. **Wails CLI o'rnatish**:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```
2. **Frontend paketlarini o'rnatish**:
   ```bash
   cd frontend
   npm install
   cd ..
   ```
3. **Dasturni yig'ish**:
   * **Windows**: `wails build`
   * **macOS**: `wails build`
   * **Linux**: `wails build` (Linuxda yig'ishdan oldin `libgtk-3-dev` va `libwebkit2gtk-4.0-dev` o'rnatilgan bo'lishi lozim).

---

## ⚙️ GitHub Actions (CI/CD) orqali Avtomatik Yig'ish
Loyiha ichiga [.github/workflows/build.yml](.github/workflows/build.yml) konfiguratsiyasi qo'shilgan.
Ushbu kodni o'zingizning GitHub repositoriyangizga push qilganingizda:
1. GitHub serverlari parallel ravishda Windows, macOS va Linux virtual mashinalarini ishga tushiradi.
2. Dastur kodlarini kompilyatsiya qiladi va har bir platforma uchun mos keladigan oflayn Pandoc faylini yuklab olib, dasturga biriktiradi (bundle).
3. Natijada sizga yuklab olish uchun tayyor oflayn paketlarni (Windows Installer `.exe`, Mac `.zip`, Linux `.tar.gz`) taqdim etadi.
