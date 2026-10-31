"""PyPasser-backed solvers — captcha types the camoufox pool can't do, done over
pure HTTP (free), plus the Friendly Captcha Go binary and a free reCAPTCHA-v3
fallback. PyPasser is imported lazily so the server still boots without it.

Every solver returns a token string (or JSON string) on success, or None on any
failure — same contract as fallback.solve_via_api, so routes/runner treat them alike.
"""
import os
import json
import asyncio

from utils.logger import logger
from utils.config import FRIENDLY_BIN


def pypasser_available() -> bool:
    try:
        import pypasser  # noqa: F401
        return True
    except ImportError:
        return False


def friendly_available() -> bool:
    return bool(FRIENDLY_BIN) and os.path.exists(FRIENDLY_BIN)


# --------------------------------------------------------------------------- PyPasser (pure HTTP)
# PyPasser's sync solvers for these types are plain httpx (no browser), so we run
# them in a worker thread to keep the asyncio loop (and camoufox) free.

async def solve_geetest(captcha_id, risk_type=None, **_):
    from pypasser.captcha import geetest

    def _run():
        kw = {"risk_type": risk_type} if risk_type else {}
        res = geetest.v4.solve(captcha_id, **kw)
        return res.response_token if res.is_solved else None   # JSON seccode string

    return await asyncio.to_thread(_run)


async def solve_lemin(captcha_id, domain, **_):
    from pypasser.captcha import lemin

    def _run():
        res = lemin.solve(captcha_id, domain=domain)
        if not res.is_solved:
            return None
        tok = res.response_token
        return json.dumps(tok) if isinstance(tok, dict) else tok

    return await asyncio.to_thread(_run)


async def solve_arcaptcha(sitekey, domain, **_):
    from pypasser.captcha import arcaptcha

    def _run():
        res = arcaptcha.solve(sitekey, domain=domain)
        return res.response_token if res.is_solved else None

    return await asyncio.to_thread(_run)


async def solve_recaptcha_v3_http(anchor_url, **_):
    """Free reCAPTCHA v3 over pure HTTP. Needs a fresh anchor URL (mine it off the
    target's reCAPTCHA iframe src). Used as the free fallback before paid CapSolver."""
    from pypasser.captcha import recaptcha

    def _run():
        res = recaptcha.v3.solve(anchor_url)
        return res.response_token if res.is_solved else None

    return await asyncio.to_thread(_run)


# --------------------------------------------------------------------------- Friendly Captcha (Go)

async def solve_friendly(sitekey, proxy=None, **_):
    """Shell out to the vendored Go PoW solver. Prints the full token on its last
    stdout line. Returns None if the binary is missing or errors."""
    if not friendly_available():
        logger.warning("Friendly binary missing; build thirdparty/friendlycaptcha -> bin/")
        return None
    args = [FRIENDLY_BIN, sitekey] + ([proxy] if proxy else [])
    try:
        proc = await asyncio.create_subprocess_exec(
            *args, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE
        )
        out, err = await asyncio.wait_for(proc.communicate(), timeout=120)
        if proc.returncode != 0:
            logger.error(f"Friendly solver exit {proc.returncode}: {err.decode(errors='ignore')[:200]}")
            return None
        lines = [l.strip() for l in out.decode(errors="ignore").splitlines() if l.strip()]
        token = lines[-1] if lines else ""
        return token or None
    except Exception as e:  # noqa: BLE001 - never crash the worker
        logger.exception(f"Friendly solver crashed: {e}")
        return None
