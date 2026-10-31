"""Camoufox (stealth Firefox) browser pool: one long-lived browser, N contexts x M pages.

Pages are recycled through an asyncio.Queue so concurrent tasks never share a page.
Contexts are rebuilt periodically to drop cookie/fingerprint buildup.
"""
import asyncio
import time
import itertools
from pathlib import Path

from camoufox import DefaultAddons
from camoufox.async_api import AsyncCamoufox

from utils.logger import logger
from utils.config import BASE_DIR, PROXY_SUPPORT, PROXY_FILE, CLEANUP_INTERVAL_MINUTES

BROWSER_ARGS = ["--no-sandbox", "--disable-setuid-sandbox"]


def load_proxies() -> list[str]:
    """Read proxies.txt (one per line). Empty / missing file -> no proxies."""
    if not PROXY_SUPPORT:
        return []
    path = Path(PROXY_FILE)
    if not path.is_absolute():
        path = BASE_DIR / PROXY_FILE
    if not path.exists():
        logger.warning(f"Proxy support on but {path} not found -> running proxyless")
        return []
    proxies = [ln.strip() for ln in path.read_text().splitlines() if ln.strip() and not ln.startswith("#")]
    logger.info(f"Loaded {len(proxies)} proxies")
    return proxies


def _parse_proxy(proxy: str) -> dict | None:
    """Accept scheme://ip:port[:user:pass] or ip:port[:user:pass]. Return Playwright proxy dict."""
    if not proxy:
        return None
    scheme = "http"
    if "://" in proxy:
        scheme, proxy = proxy.split("://", 1)
    parts = proxy.split(":")
    if len(parts) == 2:
        return {"server": f"{scheme}://{parts[0]}:{parts[1]}"}
    if len(parts) == 4:
        ip, port, user, pw = parts
        return {"server": f"{scheme}://{ip}:{port}", "username": user, "password": pw}
    logger.warning(f"Bad proxy format: {proxy} -> proxyless")
    return None


class BrowserPool:
    def __init__(self, headless: bool, thread: int, page_count: int):
        self.headless = headless
        self.thread = thread
        self.page_count = page_count
        self.queue: asyncio.Queue = asyncio.Queue()
        self.camoufox = None
        self.browser = None
        self.proxies = load_proxies()
        self._proxy_cycle = itertools.cycle(self.proxies) if self.proxies else None
        self.capacity = thread * page_count

    def _next_proxy(self) -> str | None:
        return next(self._proxy_cycle) if self._proxy_cycle else None

    async def _new_context(self):
        proxy = _parse_proxy(self._next_proxy()) if self._proxy_cycle else None
        return await self.browser.new_context(proxy=proxy) if proxy else await self.browser.new_context()

    async def start(self):
        logger.info(f"Launching camoufox (headless={self.headless}) ...")
        self.camoufox = AsyncCamoufox(
            headless=self.headless,
            exclude_addons=[DefaultAddons.UBO],
            args=BROWSER_ARGS,
        )
        self.browser = await self.camoufox.start()
        for _ in range(self.thread):
            ctx = await self._new_context()
            for _ in range(self.page_count):
                page = await ctx.new_page()
                await self.queue.put((page, ctx))
        logger.success(f"Pool ready: {self.queue.qsize()} page(s)")

    async def stop(self):
        try:
            if self.browser:
                await self.browser.close()
            if self.camoufox:
                await self.camoufox.__aexit__(None, None, None)
        except Exception as e:  # noqa: BLE001 - shutdown must never raise
            logger.warning(f"Error during browser shutdown: {e}")
        logger.success("Browser resources cleaned")

    async def acquire(self):
        return await self.queue.get()

    async def release(self, page, ctx):
        await self.queue.put((page, ctx))

    async def recycle_loop(self):
        """Every CLEANUP_INTERVAL_MINUTES, rebuild each page+context to drop state."""
        while True:
            await asyncio.sleep(CLEANUP_INTERVAL_MINUTES * 60)
            logger.info("Recycling page pool ...")
            ok = 0
            for _ in range(self.capacity):
                try:
                    page, ctx = await self.queue.get()
                    for closer in (page.close, ctx.close):
                        try:
                            await closer()
                        except Exception:  # noqa: BLE001
                            pass
                    ctx = await self._new_context()
                    page = await ctx.new_page()
                    await self.queue.put((page, ctx))
                    ok += 1
                    await asyncio.sleep(1.0)
                except Exception as e:  # noqa: BLE001
                    logger.warning(f"Recycle failed for one page: {e}")
            logger.success(f"Recycled {ok}/{self.capacity} pages")


async def reap_results_loop(results: dict, ttl: int):
    """Drop finished results older than ttl so memory doesn't grow forever."""
    while True:
        await asyncio.sleep(max(ttl, 60))
        now = time.time()
        stale = [tid for tid, r in list(results.items())
                 if r.get("status") in ("success", "error") and now - r.get("start_time", now) > ttl]
        for tid in stale:
            results.pop(tid, None)
        if stale:
            logger.debug(f"Reaped {len(stale)} stale result(s)")
