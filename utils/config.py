"""Central config. All knobs live in .env so nothing is hardcoded."""
import os
from pathlib import Path
from dotenv import load_dotenv

BASE_DIR = Path(__file__).resolve().parent.parent
load_dotenv(BASE_DIR / ".env")


def _bool(key: str, default: str = "false") -> bool:
    return os.getenv(key, default).strip().lower() in ("1", "true", "yes", "on")


# --- server ---
HOST = os.getenv("HOST", "127.0.0.1")
PORT = int(os.getenv("PORT", "5033"))

# --- browser pool ---
HEADLESS = _bool("HEADLESS", "true")
THREAD = int(os.getenv("THREAD", "1"))          # how many browser contexts
PAGE_COUNT = int(os.getenv("PAGE_COUNT", "1"))  # pages per context
PROXY_SUPPORT = _bool("PROXY_SUPPORT", "false")
PROXY_FILE = os.getenv("PROXY_FILE", "proxies.txt")
CLEANUP_INTERVAL_MINUTES = int(os.getenv("CLEANUP_INTERVAL_MINUTES", "60"))

# --- task limits ---
SOLVE_TIMEOUT = int(os.getenv("SOLVE_TIMEOUT", "60"))     # per-task hard cap (s)
RESULT_TTL = int(os.getenv("RESULT_TTL", "300"))          # keep result this long (s)

# --- paid-API fallback (optional, local-first) ---
FALLBACK_ENABLED = _bool("FALLBACK_ENABLED", "false")
CAPSOLVER_API_KEY = os.getenv("CAPSOLVER_API_KEY", "").strip()
CAPSOLVER_API = os.getenv("CAPSOLVER_API", "https://api.capsolver.com")

# --- logging ---
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")
LOG_ROTATION = os.getenv("LOG_ROTATION", "10 MB")
LOG_RETENTION = os.getenv("LOG_RETENTION", "7 days")

STATIC_DIR = BASE_DIR / "static"
LOGS_DIR = BASE_DIR / "logs"
