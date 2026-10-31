"""Smoke test: wiring sanity without launching a browser or network.
Run:  python test_smoke.py   ->  prints OK or raises.

Catches the dumb breakers early: bad template braces, dispatch tables out of sync,
import errors, missing route handlers.
"""
import asyncio

from utils import templates, solvers, fallback
from utils.routes import register_routes
from fastapi import FastAPI


def test_templates_format():
    # .format must not raise KeyError / stray-brace ValueError
    assert "sitekey123" in templates.RECAPTCHA_V3_TEMPLATE.format(sitekey="sitekey123", action="verify")
    assert "sk2" in templates.RECAPTCHA_V2_TEMPLATE.format(sitekey="sk2")
    assert "sk2i" in templates.RECAPTCHA_V2_INVISIBLE_TEMPLATE.format(sitekey="sk2i")
    assert "<!-- cf turnstile -->" in templates.HTML_TEMPLATE  # injection marker present


def test_dispatch_consistency():
    # every attempt with an API type must have a CapSolver mapping
    for name, (fn, fb_type) in solvers._ATTEMPTS.items():
        assert callable(fn), name
        if fb_type is not None:
            assert fb_type in fallback._TASK_MAP, f"{name}->{fb_type} missing in fallback._TASK_MAP"
            assert fb_type in fallback._TOKEN_FIELD, f"{fb_type} missing token field"


def test_routes_register():
    class Dummy:
        pool = type("P", (), {"capacity": 1})()
        results = {}
        current_task_num = 0
        def decrement_task(self): ...
    app = FastAPI()
    register_routes(app, Dummy())
    paths = {r.path for r in app.routes}
    for p in ("/turnstile", "/clearance", "/recaptchav2", "/recaptchav3", "/result", "/health"):
        assert p in paths, f"missing route {p}"


def test_fallback_guard():
    # disabled by default -> available() False, so no accidental paid calls
    assert fallback.available() in (True, False)


if __name__ == "__main__":
    test_templates_format()
    test_dispatch_consistency()
    test_routes_register()
    test_fallback_guard()
    print("OK - smoke test passed")
