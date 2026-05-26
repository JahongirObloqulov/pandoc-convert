# Pandoc Pro Konverteri

Pandoc Pro — bu Markdown, Microsoft Word (DOCX), PDF, HTML, EPUB, LaTeX va oddiy matnli (TXT) hujjatlarni o'zaro bir formatdan ikkinchisiga oson va professional darajada o'giruvchi cross-platform desktop ilova. 

Ushbu ilova **Wails (Go + HTML/JS/CSS)** texnologiyasi asosida yasalgan bo'lib, o'z ichida oflayn ishlash uchun moslashtirilgan milliy va xalqaro formatlarni qo'llab-quvvatlaydi.

---

## 🚀 Asosiy Imkoniyatlar
* **Premium Glassmorphic Interfeys**: Zamonaviy to'q fon (Dark Mode), shaffof dizayn (glassmorphism) va silliq animatsiyalar.
* **Mavzular (Night/Light Mode)**: Tungi (Dark) va yorug' (Light) interfeys rejimlari. Foydalanuvchi tanlagan rejim saqlanib qoladi (localStorage yordamida) va keyingi safar avtomatik yuklanadi.
* **Ko'p Tilli Interfeys (i18n)**: O'zbek (UZ), Rus (RU) va Ingliz (EN) tillarini to'liq qo'llab-quvvatlash. Ilova ishga tushganda tizim tilini avtomatik aniqlaydi, shuningdek interfeysdagi pill-buttonlar orqali tilni darhol o'zgartirish mumkin.
* **Drag & Drop**: Fayllarni interfeysga shunchaki sudrab tashlash orqali yuklash.
* **Fayl Preview**: Matnli hujjatlar (Markdown, TXT, HTML, LaTeX) tarkibini o'girishdan oldin dastur ichida tezkor ko'rish.
* **Avtomatik Taklif (Auto-suggest)**: Asl fayl tanlanganda chiqish fayli nomi va saqlash joyini avtomatik to'ldirish.
* **Tezkor Harakatlar**: Fayl o'girilgandan so'ng uni to'g'ridan-to'g'ri standart dastur orqali ochish yoki u joylashgan papkani Windows Explorer/macOS Finder orqali belgilangan holda ko'rish.
* **Tarix Paneli**: Oxirgi 5 ta o'girilgan fayllarni tezkor kuzatish va ularni tarixdan turib ham ochish.
* **100% Oflayn Ishlash**: Dastur ishlashi uchun kompyuterda internet bo'lishi yoki tizimga Pandoc dasturini alohida o'rnatish talab etilmaydi (chunki oflayn paket ichida kerakli barcha vositalar jamlangan).

---

## 📦 Oflayn Dasturni O'rnatish va Ishga Tushirish

### 1. Windows (PC)
1. GitHub Actions'dan yuklab olingan [pandoc-pro-installer.exe](build/bin/pandoc-pro-installer.exe) o'rnatuvchi faylini ishga tushiring.
2. Dastur avtomatik tarzda `C:\Program Files\PandocConvertPro` papkasiga o'rnatiladi, tizimga oflayn `pandoc.exe` ni joylaydi va ish stolida yorliq hosil qiladi.
3. Dasturni ishga tushiring va oflayn rejimda ishlatishingiz mumkin.

### 2. macOS (MacBook Intel & M1/M2/M3/M4)
macOS xavfsizlik qoidalari tufayli imzolanmagan (unsigned) dasturlarni to'g'ridan-to'g'ri ochishni bloklashi mumkin. Quyidagi bosqichlarni bajaring:
1. `pandoc-pro-mac.zip` faylini arxivdan oching. Ichidan `pandoc-app.app` dasturi chiqadi.
2. Dasturni `Applications` (Dasturlar) papkasiga ko'chiring.
3. Terminal dasturini oching va quyidagi buyruqni kiriting:
   ```bash
   xattr -cr /Applications/pandoc-app.app
   ```
4. Dasturni odatiy tarzda oching. Ilova ichiga macOS uchun maxsus moslashtirilgan oflayn `pandoc` fayli joylashtirilgani sababli, u internetsiz to'liq ishlaydi.

### 3. Linux (Ubuntu / Debian / Fedora va boshqalar)
1. `pandoc-pro-linux.tar.gz` faylini istalgan papkaga arxivdan chiqaring.
2. Arxiv ichidan `pandoc-app` va `pandoc` fayllari chiqadi (ular bitta papkada turishi shart).
3. Terminal orqali dasturga ishga tushish huquqini bering:
   ```bash
   chmod +x pandoc-app pandoc
   ```
4. Dasturni ishga tushiring:
   ```bash
   ./pandoc-app
   ```

---

## 🛠️ Loyihani Mahalliy Yig'ish (Local Compilation)

Agar dasturni o'zingiz mahalliy kompyuteringizda noldan yig'moqchi bo'lsangiz, kompyuteringizda **Go (1.23+)** va **Node.js (18+)** o'rnatilgan bo'lishi talab etiladi.

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
   * **Windows uchun**:
     ```bash
     wails build -nopackage
     ```
     *(Keyin `installer.nsi` skriptini kompilyatsiya qilib shaxsiy installer yaratishingiz mumkin).*
   * **macOS uchun**:
     ```bash
     wails build
     ```
   * **Linux uchun (Ubuntu 22.04 va undan eski versiyalar)**:
     ```bash
     sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev build-essential pkg-config
     wails build
     ```
   * **Linux uchun (Ubuntu 23.04, 24.04+ va yangi distributivlar)**:
     Ubuntu 24.04 da eski `webkit2gtk-4.0` o'rniga yangi `webkit2gtk-4.1` ishlatiladi. Yig'ish quyidagicha bo'ladi:
     ```bash
     sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev build-essential pkg-config
     wails build -tags webkit2_41
     ```

---

## ⚙️ GitHub Actions (CI/CD) orqali Avtomatik Yig'ish
Loyiha ichiga [.github/workflows/build.yml](.github/workflows/build.yml) konfiguratsiyasi qo'shilgan.
Ushbu kodni o'zingizning GitHub repositoriyangizga push qilganingizda:
1. GitHub serverlari parallel ravishda Windows, macOS va Linux virtual mashinalarini ishga tushiradi.
2. Dastur kodlarini kompilyatsiya qiladi va har bir platforma uchun mos keladigan oflayn Pandoc faylini yuklab olib, dasturga biriktiradi (bundle).
   * **Windows** da NSIS dasturi avtomatik o'rnatilib, installer yig'iladi.
   * **Linux (Ubuntu 24.04)** da kompilyatsiya `webkit2_41` tagi bilan bajariladi.
3. Natijada sizga yuklab olish uchun tayyor oflayn paketlarni (Windows Installer `.exe`, Mac `.zip`, Linux `.tar.gz`) taqdim etadi.
