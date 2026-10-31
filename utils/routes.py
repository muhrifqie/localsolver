"""HTTP API. Submit a solve -> get task_id (202). Poll /result?id=... until 200/4xx/5xx.

One generic `_submit` kills the per-endpoint copy-paste the reference repos had.
"""
import time
import uuid
import asyncio

from fastapi import Query
from fastapi.responses import JSONResponse, FileResponse

from utils.logger import logger
from utils.solvers import run_task, run_http_task
from utils.config import RESULT_TTL, STATIC_DIR, HTTP_CAPACITY
from utils import fallback, pypasser_solvers


def _submit(server, name: str, required: dict, **params):
    """Validate, enforce capacity, spawn the task. Returns a JSONResponse."""
    missing = [k for k, v in required.items() if not v]
    if missing:
        return JSONResponse({"status": "error", "error": f"Required: {', '.join(missing)}"}, status_code=400)
    if server.current_task_num >= server.pool.capacity:
        return JSONResponse({"status": "error", "error": "Server at max capacity"}, status_code=429)

    task_id = str(uuid.uuid4())
    server.current_task_num += 1
    logger.info(f"New {name} task: {task_id}")
    asyncio.create_task(
        run_task(name, task_id, server.results, server.pool, server.decrement_task, **params)
    )
    return JSONResponse({"task_id": task_id, "status": "accepted"}, status_code=202)


def _submit_http(server, name: str, required: dict, solver, **params):
    """Spawn a pool-less (PyPasser/Go) solve. Capacity gated separately from the browser pool."""
    missing = [k for k, v in required.items() if not v]
    if missing:
        return JSONResponse({"status": "error", "error": f"Required: {', '.join(missing)}"}, status_code=400)
    if server.http_task_num >= HTTP_CAPACITY:
        return JSONResponse({"status": "error", "error": "Server at max capacity"}, status_code=429)
    task_id = str(uuid.uuid4())
    server.http_task_num += 1
    logger.info(f"New {name} task: {task_id}")
    asyncio.create_task(
        run_http_task(name, task_id, server.results, solver, server.decrement_http, **params)
    )
    return JSONResponse({"task_id": task_id, "status": "accepted"}, status_code=202)


def register_routes(app, server):

    @app.get("/", include_in_schema=False)
    async def console():
        return FileResponse(STATIC_DIR / "index.html")

    @app.get("/health")
    async def health():
        return {"status": "ok", "capacity": server.pool.capacity,
                "in_flight": server.current_task_num, "http_in_flight": server.http_task_num,
                "fallback": fallback.available(),
                "pypasser_free": pypasser_solvers.pypasser_available(),
                "friendly": pypasser_solvers.friendly_available()}

    @app.get("/turnstile")
    async def turnstile(url: str = Query(...), sitekey: str = Query(...),
                        action: str = Query(None), cdata: str = Query(None)):
        return _submit(server, "turnstile", {"url": url, "sitekey": sitekey},
                       url=url, sitekey=sitekey, action=action, cdata=cdata)

    @app.get("/clearance")
    async def clearance(url: str = Query(...)):
        return _submit(server, "clearance", {"url": url}, url=url)

    @app.get("/recaptchav2")
    async def recaptcha_v2(url: str = Query(...), sitekey: str = Query(...)):
        return _submit(server, "recaptcha_v2", {"url": url, "sitekey": sitekey},
                       url=url, sitekey=sitekey, invisible=False)

    @app.get("/recaptchav2invisible")
    async def recaptcha_v2_invisible(url: str = Query(...), sitekey: str = Query(...)):
        return _submit(server, "recaptcha_v2", {"url": url, "sitekey": sitekey},
                       url=url, sitekey=sitekey, invisible=True)

    @app.get("/recaptchav3")
    async def recaptcha_v3(url: str = Query(...), sitekey: str = Query(...),
                           action: str = Query("verify"), min_score: float = Query(None),
                           anchor_url: str = Query(None)):
        # anchor_url (optional): enables the free PyPasser HTTP fallback if the browser solve fails
        return _submit(server, "recaptcha_v3", {"url": url, "sitekey": sitekey},
                       url=url, sitekey=sitekey, action=action, min_score=min_score, anchor_url=anchor_url)

    # --- free PyPasser / Go solvers (pure HTTP, no browser) ---
    @app.get("/recaptchav3http")
    async def recaptcha_v3_http(anchor_url: str = Query(...)):
        return _submit_http(server, "recaptcha_v3_http", {"anchor_url": anchor_url},
                            pypasser_solvers.solve_recaptcha_v3_http, anchor_url=anchor_url)

    @app.get("/geetest")
    async def geetest(captcha_id: str = Query(...), risk_type: str = Query(None)):
        return _submit_http(server, "geetest_v4", {"captcha_id": captcha_id},
                            pypasser_solvers.solve_geetest, captcha_id=captcha_id, risk_type=risk_type)

    @app.get("/lemin")
    async def lemin(captcha_id: str = Query(...), domain: str = Query(...)):
        return _submit_http(server, "lemin", {"captcha_id": captcha_id, "domain": domain},
                            pypasser_solvers.solve_lemin, captcha_id=captcha_id, domain=domain)

    @app.get("/arcaptcha")
    async def arcaptcha(sitekey: str = Query(...), domain: str = Query(...)):
        return _submit_http(server, "arcaptcha", {"sitekey": sitekey, "domain": domain},
                            pypasser_solvers.solve_arcaptcha, sitekey=sitekey, domain=domain)

    @app.get("/friendlycaptcha")
    async def friendlycaptcha(sitekey: str = Query(...), proxy: str = Query(None)):
        return _submit_http(server, "friendly", {"sitekey": sitekey},
                            pypasser_solvers.solve_friendly, sitekey=sitekey, proxy=proxy)

    @app.get("/result")
    async def result(task_id: str = Query(..., alias="id")):
        res = server.results.get(task_id)
        if not res:
            return JSONResponse({"status": "error", "message": "Invalid or expired task_id"}, status_code=404)

        if res.get("status") == "process":
            if time.time() - res.get("start_time", time.time()) > RESULT_TTL:
                res = {"status": "error", "value": "timeout", "message": "Task timeout"}
                server.results[task_id] = res
            else:
                return JSONResponse(res, status_code=202)  # still working

        res = server.results.pop(task_id, res)  # terminal state -> hand out once
        if res.get("status") == "success":
            code = 200
        elif res.get("value") == "timeout":
            code = 408
        elif str(res.get("value", "")).endswith("_fail"):
            code = 422
        else:
            code = 500
        return JSONResponse(res, status_code=code)
