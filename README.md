# localsolver

Server **solver captcha lokal** — jalan di mesin sendiri, pakai browser siluman
**camoufox** + pool halaman, dibungkus API HTTP (FastAPI). Local-first: nyoba
nyelesaiin captcha pakai browser dulu (gratis), baru fallback ke **CapSolver**
(API bayar) kalau lokal gagal dan fallback diaktifin.

Dibangun dari pelajaran 5 repo acuan — detail tiap repo + pola error & cara fix
ada di [`NOTES.md`](NOTES.md).

## Yang didukung
| Endpoint | Captcha | Param wajib |
|---|---|---|
| `GET /turnstile` | Cloudflare Turnstile | `url`, `sitekey` (opsional `action`, `cdata`) |
| `GET /clearance` | Cookie `cf_clearance` (lewat challenge CF) | `url` |
| `GET /recaptchav2` | reCAPTCHA v2 | `url`, `sitekey` |
| `GET /recaptchav2invisible` | reCAPTCHA v2 invisible | `url`, `sitekey` |
| `GET /recaptchav3` | reCAPTCHA v3 | `url`, `sitekey` (opsional `action`, `min_score`) |
| `GET /result?id=<task_id>` | Ambil hasil | `id` |
| `GET /health` | Status server | — |

Pola pakai = **async**: submit → dapat `task_id` (202) → poll `/result?id=...`
sampai `200` (sukses), `422` (gagal solve), `408` (timeout), atau `404` (expired).

## Jalanin (Windows)
```powershell
cd C:\BOT\localsolver
.\run.ps1
```
Sekali jalan pertama: bikin venv, install deps, download browser camoufox, copy
`.env`. Edit `.env` sesuai kebutuhan (lihat `.env.example`).

Manual (kalau nggak pakai run.ps1):
```powershell
python -m venv .venv; .\.venv\Scripts\Activate.ps1
pip install -r requirements.txt
python -m camoufox fetch
copy .env.example .env
python main.py
```

## Contoh
```bash
# submit
curl "http://127.0.0.1:5033/turnstile?url=https://target.com&sitekey=0x4AAA..."
# -> {"task_id":"abc...","status":"accepted"}

# poll
curl "http://127.0.0.1:5033/result?id=abc..."
# -> {"status":"success","value":"<token>","source":"local","elapsed_time":3.1}
```

## Konfigurasi penting (`.env`)
- `HEADLESS=false` → lihat browser-nya buat debug satu situs.
- `THREAD` × `PAGE_COUNT` = jumlah task paralel. Naikin kalau butuh throughput (lebih berat RAM).
- `FALLBACK_ENABLED=true` + `CAPSOLVER_API_KEY=...` → nyalain jaring pengaman API. Default mati (100% lokal/gratis).
- `PROXY_SUPPORT=true` + isi `proxies.txt` → tiap context pakai proxy (rotasi).

## Catatan
- **Target = aset/izin owner.** Tool riset keamanan, bukan buat nyerang sistem orang.
- Log lengkap (termasuk stack error) di `logs/solver_<tanggal>.log` — ini sumber utama buat debug kalau ada yang gagal.
- Smoke test cepat tanpa browser: `python test_smoke.py`.
