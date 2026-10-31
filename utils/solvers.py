"""Browser-based (local) solvers + a DRY runner that handles timing, pooling,
timeout, result writing and the optional API fallback.

Each captcha type = one small `_attempt_*` coroutine that drives a page and returns a
token/clearance string or None. `run_task` wraps them all with the shared plumbing.
"""
import time
import asyncio

from utils.logger import logger
from utils.config import SOLVE_TIMEOUT
from utils import fallback
from utils.templates import (
    HTML_TEMPLATE,
    RECAPTCHA_V3_TEMPLATE,
    RECAPTCHA_V2_TEMPLATE,
    RECAPTCHA_V2_INVISIBLE_TEMPLATE,
)


def _with_slash(url: str) -> str:
    return url if url.endswith("/") else url + "/"


async def _serve(page, url: str, html: str):
    """Intercept `url` and return our own HTML so the widget runs under the real origin."""
    await page.route(url, lambda route: route.fulfill(body=html, status=200,
                                                       content_type="text/html"))
    await page.goto(url)


# ----------------------------------------------------------------------------- attempts

async def _attempt_turnstile(page, url, sitekey, action=None, cdata=None):
    url = _with_slash(url)
    div = (f'<div class="cf-turnstile" data-sitekey="{sitekey}"'
           + (f' data-action="{action}"' if action else "")
           + (f' data-cdata="{cdata}"' if cdata else "")
           + "></div>")
    html = HTML_TEMPLATE.replace("<!-- cf turnstile -->", div)
    await _serve(page, url, html)
    for _ in range(60):
        try:
            val = await page.input_value("[name=cf-turnstile-response]", timeout=250)
            if val:
                return val
            # nudge the widget if it rendered an interactive checkbox
            try:
                await page.locator(".cf-turnstile").click(timeout=250)
            except Exception:  # noqa: BLE001
                pass
        except Exception:  # noqa: BLE001 - element not ready yet
            pass
        await asyncio.sleep(0.1)
    return None


async def _attempt_recaptcha_v3(page, url, sitekey, action="verify", min_score=None):
    url = _with_slash(url)
    await _serve(page, url, RECAPTCHA_V3_TEMPLATE.format(sitekey=sitekey, action=action))
    for _ in range(60):
        try:
            token = await page.input_value("#recaptcha-token", timeout=250)
            if token:
                return token
        except Exception:  # noqa: BLE001
            pass
        await asyncio.sleep(0.1)
    return None


async def _attempt_recaptcha_v2(page, url, sitekey, invisible=False):
    url = _with_slash(url)
    tmpl = RECAPTCHA_V2_INVISIBLE_TEMPLATE if invisible else RECAPTCHA_V2_TEMPLATE
    await _serve(page, url, tmpl.format(sitekey=sitekey))
    await asyncio.sleep(1)
    for _ in range(80):
        try:
            token = await page.input_value("#recaptcha-response", timeout=300)
            if token:
                return token
            if not invisible:
                # try to tick the "I'm not a robot" checkbox inside its iframe
                try:
                    frame_el = await page.query_selector("iframe[title*='reCAPTCHA']")
                    if frame_el:
                        frame = await frame_el.content_frame()
                        if frame:
                            await frame.click(".recaptcha-checkbox-border", timeout=300)
                except Exception:  # noqa: BLE001
                    pass
        except Exception:  # noqa: BLE001
            pass
        await asyncio.sleep(0.2)
    return None


async def _attempt_clearance(page, url, **_):
    """Load the real protected URL, let Cloudflare's managed challenge pass, grab cf_clearance."""
    await page.goto(url, wait_until="domcontentloaded")
    for _ in range(60):
        cookies = await page.context.cookies()
        for c in cookies:
            if c["name"] == "cf_clearance" and c.get("value"):
                return c["value"]
        await asyncio.sleep(0.5)
    return None


# dispatch table: name -> (attempt coroutine, captcha_type used for API fallback)
_ATTEMPTS = {
    "turnstile": (_attempt_turnstile, "turnstile"),
    "recaptcha_v3": (_attempt_recaptcha_v3, "recaptcha_v3"),
    "recaptcha_v2": (_attempt_recaptcha_v2, "recaptcha_v2"),
    "clearance": (_attempt_clearance, None),  # no public API equivalent -> local only
}


# ----------------------------------------------------------------------------- runner

async def run_task(name: str, task_id: str, results: dict, pool, decrement, **params):
    """Shared plumbing for every captcha type: pool, timeout, fallback, result write."""
    attempt, fb_type = _ATTEMPTS[name]
    start = time.time()
    results[task_id] = {"status": "process", "message": f"solving {name}", "start_time": start}
    page, ctx = await pool.acquire()
    token = None
    try:
        token = await asyncio.wait_for(attempt(page, **params), timeout=SOLVE_TIMEOUT)
    except asyncio.TimeoutError:
        logger.warning(f"{name} hard-timeout, task {task_id}")
    except Exception as e:  # noqa: BLE001 - one bad solve must not kill the worker
        logger.exception(f"{name} crashed, task {task_id}: {e}")
    finally:
        try:
            await page.unroute_all(behavior="ignoreErrors")
        except Exception:  # noqa: BLE001
            pass
        await pool.release(page, ctx)
        decrement()

    source = "local"
    if not token and fb_type and fallback.available():
        logger.info(f"{name} local failed -> trying CapSolver fallback, task {task_id}")
        token = await fallback.solve_via_api(fb_type, **params)
        source = "capsolver"

    elapsed = round(time.time() - start, 3)
    if token:
        results[task_id] = {"status": "success", "value": token, "source": source,
                            "elapsed_time": elapsed, "start_time": start}
        logger.success(f"{name} solved via {source}, task {task_id}, {elapsed}s")
    else:
        results[task_id] = {"status": "error", "value": f"{name}_fail",
                            "elapsed_time": elapsed, "start_time": start}
        logger.warning(f"{name} failed, task {task_id}, {elapsed}s")
