# NOTES — basis pengetahuan solver (buat debug & fix)

Rangkuman 5 repo acuan + jebakan yang udah kelihatan, biar kalau localsolver error
kita tahu akar & cara benerinnya. Diupdate tiap nemu kasus baru.

---

## Repo acuan & apa yang diambil

### 1. `DenimEvert/camoufox-capsolver`
- **Pola:** camoufox (browser) + CapSolver (API bayar) buat Turnstile.
- **API CapSolver:** `POST /createTask` → `taskId`; poll `POST /getTaskResult` tiap 1 dtk sampai `status=="ready"`, token di `solution.token`. Error ditandai `errorId != 0`.
- **Diambil ke:** `utils/fallback.py` (persis pola create/poll ini).
- **Jebakan:** token Turnstile harus di-*inject* balik ke `input[name=cf-turnstile-response]` lalu panggil callback `data-callback` — kalau nggak, form nolak walau token valid.

### 2. `techinz/playwright-captcha`
- **Pola:** library pip, OO (`BaseSolver`, `ClickSolver`, `ApiSolverBase`), alur **detect → solve → apply → submit**. Dukung Turnstile, CF Interstitial, reCAPTCHA v2/v3. Backend: ClickSolver (browser siluman) atau API (2captcha/10captcha).
- **Diambil:** konsep "click solver pakai kesiluman browser" (gratis) + arsitektur solver yang bisa ditukar backend-nya (local ↔ API).
- **Jebakan BESAR (camoufox):** `add_init_script` camoufox lagi bermasalah → butuh workaround addon `techinz/camoufox-add_init_script`, plus flag `main_world_eval=True`, `config={'forceScopeAccess': True}`, `disable_coop=True`, `i_know_what_im_doing=True` saat bikin instance camoufox. **Kalau nanti kita pindah ke solver berbasis `add_init_script`, wajib pasang ini.** (localsolver sekarang nggak pakai `add_init_script`, jadi aman.)

### 3. `najibyahya/Boterdrop-Solver`
- **Pola:** FastAPI + pool camoufox. Endpoint `/turnstile`, `/clearance` (cf_clearance), `/aws-token` (AWS WAF), `/recaptchaV3`. Ada dukungan proxy + cleanup berkala.
- **Diambil:** fix **Windows event loop** (`WindowsSelectorEventLoopPolicy`) — WAJIB di Windows, kalau nggak camoufox/subprocess gagal start. Konsep endpoint `/clearance` & `/aws-token`.
- **Jebakan:** `headless` default `true`; sebagian challenge CF cuma lolos kalau fingerprint + IP bagus (proxy residensial). cf_clearance **terikat IP + User-Agent** — token nggak kepake kalau dipakai dari IP/UA beda.

### 4. `agathasangkara/solver`  ← **backbone utama**
- **Pola:** FastAPI modular (`utils/config,logger,browser,solvers,routes,templates`), loguru, pool `asyncio.Queue`, submit→`task_id`→poll `/result`. Trik inti: `page.route(url, fulfill(html))` → serve HTML sendiri di origin target, widget jalan, token dibaca dari input tersembunyi.
- **Diambil:** struktur folder, pool + recycle context, template HTML reCAPTCHA v2/v3, pola polling token, kode status `/result` (200/408/422/404).
- **Jebakan:** route handler numpuk di page yang di-recycle → localsolver nambah `page.unroute_all()` di `finally`. reCAPTCHA v2 non-invisible kadang perlu klik checkbox di dalam iframe.

### 5. `agathasangkara/solver` modul `arkoselab.py`
- FunCaptcha / Arkose Labs. **Belum** diimplement di localsolver (butuh handling iframe Arkose + public key). Catatan buat nanti kalau ada target Arkose.

---

## Arsitektur localsolver (ringkas)
```
main.py            -> FastAPI app + lifecycle + Windows loop fix
utils/config.py    -> semua env (.env)
utils/logger.py    -> loguru console + file (backtrace ON)
utils/browser.py   -> BrowserPool (camoufox, context/page queue, recycle, proxy)
utils/templates.py -> HTML widget per captcha
utils/solvers.py   -> _attempt_* (local) + run_task (plumbing + fallback)
utils/fallback.py  -> CapSolver (dipakai cuma kalau lokal gagal & diaktifin)
utils/routes.py    -> endpoint HTTP (submit async + /result)
```
**Local-first:** `run_task` nyoba browser dulu; gagal + `FALLBACK_ENABLED` → CapSolver.
Field `source` di hasil nunjukin `local` atau `capsolver`.

---

## Pola error umum & cara fix

| Gejala | Akar | Fix |
|---|---|---|
| `AttributeError: 'FastAPI' object has no attribute 'add_event_handler'` | FastAPI/Starlette baru hapus `add_event_handler` (acuan pakai versi lama) | Pakai `lifespan` context manager (udah diterapkan di `main.py`) — **jangan** balik ke `add_event_handler` |
| Server nggak start di Windows (`NotImplementedError` subprocess) | event loop salah | Udah ada `WindowsSelectorEventLoopPolicy` di `main.py` — pastikan nggak kehapus |
| `camoufox` error saat start / browser nggak ketemu | browser belum di-fetch | `python -m camoufox fetch` |
| Semua solve `*_fail` padahal sitekey bener | halaman/widget nggak render (headless kedetek, atau origin salah) | set `HEADLESS=false` buat lihat; cek `url` bener (origin target, bukan URL endpoint widget) |
| Turnstile dapat token tapi situs target nolak | token nggak di-inject + callback nggak dipanggil | di sisi pemakai: isi `cf-turnstile-response` lalu panggil `data-callback` (lihat repo #1) |
| cf_clearance nggak kepake di request lain | cf_clearance terikat IP+UA | pakai IP & User-Agent yang SAMA saat pakai cookie-nya |
| `429 Server at max capacity` | pool penuh | naikin `THREAD`/`PAGE_COUNT`, atau antre request |
| reCAPTCHA v3 skor rendah (ditolak server) | reputasi IP/fingerprint jelek | pakai proxy residensial bagus; v3 soal skor, bukan cuma token |
| Memory naik terus | result numpuk | udah ada `reap_results_loop` + `RESULT_TTL`; cek kalau diubah |
| Route handler makin lambat tiap request | `page.route` numpuk di page recycled | udah di-`unroute_all` di `finally` `run_task` |
| CapSolver `errorId` / saldo habis | fallback | cek log `CapSolver ... error`; isi/isi ulang key, atau matikan `FALLBACK_ENABLED` |

**Aturan debug:** selalu baca `logs/solver_<tanggal>.log` dulu (stack lengkap karena
`backtrace/diagnose` ON). Reproduksi dengan `HEADLESS=false` buat lihat apa yang
sebenarnya terjadi di browser.

## Status uji coba (31 Okt 2026)
Diuji pakai test sitekey resmi (CF `1x..AA/BB/2x..AB`, Google v2 `6LeIxAcT...`):
Turnstile visible ✅ (1.9s) · Turnstile invisible ✅ (1.7s) · Turnstile always-block ✅ gagal benar (422) ·
reCAPTCHA v2 ✅ (2.5s) · v2 invisible ✅ (1.4s) · v3 ✅ token (0.8s). Semua via `local`, fallback mati.
Belum: `/clearance` (butuh target ber-challenge Cloudflare milik owner).

## TODO / belum diuji
- [ ] `/clearance` di target CF milik owner (butuh situs yang munculin managed challenge).
- [ ] AWS WAF `/aws-token` (ada di Boterdrop, belum di-port).
- [ ] Arkose/FunCaptcha (ada di agathasangkara, belum di-port).
- [ ] Verifikasi workaround `add_init_script` kalau pindah ke solver berbasis itu.
