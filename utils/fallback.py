"""Optional paid-API fallback (CapSolver). Local-first: only called when the browser
solve fails AND FALLBACK_ENABLED=true with a key set. Keeps the server usable on sites
the local click-solver can't beat, without paying when it can.

Extend _TASK_MAP to add providers (2captcha etc.) or captcha types later.
"""
import asyncio
import httpx

from utils.logger import logger
from utils.config import FALLBACK_ENABLED, CAPSOLVER_API_KEY, CAPSOLVER_API


def available() -> bool:
    return FALLBACK_ENABLED and bool(CAPSOLVER_API_KEY)


# captcha_type -> function building the CapSolver task payload
def _turnstile_task(url, sitekey, **_):
    return {"type": "AntiTurnstileTaskProxyLess", "websiteURL": url, "websiteKey": sitekey}


def _recaptcha_v2_task(url, sitekey, **_):
    return {"type": "ReCaptchaV2TaskProxyLess", "websiteURL": url, "websiteKey": sitekey}


def _recaptcha_v3_task(url, sitekey, action=None, min_score=0.3, **_):
    return {"type": "ReCaptchaV3TaskProxyLess", "websiteURL": url, "websiteKey": sitekey,
            "pageAction": action or "verify", "minScore": min_score}


_TASK_MAP = {
    "turnstile": _turnstile_task,
    "recaptcha_v2": _recaptcha_v2_task,
    "recaptcha_v2_invisible": _recaptcha_v2_task,
    "recaptcha_v3": _recaptcha_v3_task,
}

# which solution field carries the token, per CapSolver type
_TOKEN_FIELD = {"turnstile": "token", "recaptcha_v2": "gRecaptchaResponse",
                "recaptcha_v2_invisible": "gRecaptchaResponse", "recaptcha_v3": "gRecaptchaResponse"}


async def solve_via_api(captcha_type: str, **kwargs) -> str | None:
    """Return a token string, or None on any failure (caller already logged the local fail)."""
    builder = _TASK_MAP.get(captcha_type)
    if not builder:
        logger.warning(f"No CapSolver mapping for {captcha_type}")
        return None
    payload = builder(**kwargs)
    try:
        async with httpx.AsyncClient(timeout=30) as client:
            r = await client.post(f"{CAPSOLVER_API}/createTask",
                                  json={"clientKey": CAPSOLVER_API_KEY, "task": payload})
            data = r.json()
            if data.get("errorId"):
                logger.error(f"CapSolver createTask error: {data.get('errorDescription')}")
                return None
            task_id = data["taskId"]

            for _ in range(60):  # ~60s cap
                await asyncio.sleep(1)
                r = await client.post(f"{CAPSOLVER_API}/getTaskResult",
                                      json={"clientKey": CAPSOLVER_API_KEY, "taskId": task_id})
                data = r.json()
                if data.get("status") == "ready":
                    field = _TOKEN_FIELD.get(captcha_type, "token")
                    return data.get("solution", {}).get(field)
                if data.get("status") == "failed" or data.get("errorId"):
                    logger.error(f"CapSolver task failed: {data.get('errorDescription')}")
                    return None
            logger.error("CapSolver timed out")
            return None
    except Exception as e:  # noqa: BLE001 - fallback must never crash the server
        logger.exception(f"CapSolver request crashed: {e}")
        return None
