"""localsolver entrypoint.

Run:  python main.py
Local-first captcha solver API backed by a camoufox browser pool, with optional
CapSolver fallback. See README.md for endpoints and NOTES.md for how each piece works.
"""
import sys
import asyncio
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
import uvicorn

from utils.config import (HOST, PORT, HEADLESS, THREAD, PAGE_COUNT, RESULT_TTL, STATIC_DIR)
from utils.logger import logger
from utils.browser import BrowserPool, reap_results_loop
from utils.routes import register_routes

# Windows + asyncio subprocess (camoufox/playwright) needs the selector loop.
if sys.platform == "win32":
    asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())


class SolverServer:
    def __init__(self):
        self.pool = BrowserPool(HEADLESS, THREAD, PAGE_COUNT)
        self.results: dict = {}
        self.current_task_num = 0
        self.http_task_num = 0   # pool-less (PyPasser/Go) solves

        STATIC_DIR.mkdir(parents=True, exist_ok=True)
        self.app = FastAPI(title="localsolver", version="0.1.0", lifespan=self._lifespan)
        self.app.mount("/static", StaticFiles(directory=str(STATIC_DIR)), name="static")
        register_routes(self.app, self)

    def decrement_task(self):
        self.current_task_num = max(0, self.current_task_num - 1)

    def decrement_http(self):
        self.http_task_num = max(0, self.http_task_num - 1)

    @asynccontextmanager
    async def _lifespan(self, app):
        # startup
        try:
            await self.pool.start()
        except Exception as e:  # noqa: BLE001
            logger.exception(f"Browser init failed: {e}")
            raise
        asyncio.create_task(self.pool.recycle_loop())
        asyncio.create_task(reap_results_loop(self.results, RESULT_TTL))
        logger.success(f"localsolver up on http://{HOST}:{PORT}")
        yield
        # shutdown
        await self.pool.stop()


server = SolverServer()
app = server.app


if __name__ == "__main__":
    uvicorn.run(app, host=HOST, port=PORT, log_level="warning")
