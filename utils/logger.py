"""loguru logger: colored console + daily rotating file, so every failure is traceable."""
import sys
from loguru import logger
from utils.config import LOGS_DIR, LOG_LEVEL, LOG_ROTATION, LOG_RETENTION

LOGS_DIR.mkdir(parents=True, exist_ok=True)

logger.remove()
logger.add(
    sys.stderr,
    level=LOG_LEVEL,
    format="<green>{time:YYYY-MM-DD HH:mm:ss}</green> | <level>{level: <8}</level> | "
           "<cyan>{function}</cyan>:<cyan>{line}</cyan> - <level>{message}</level>",
)
logger.add(
    str(LOGS_DIR / "solver_{time:YYYY-MM-DD}.log"),
    level=LOG_LEVEL,
    rotation=LOG_ROTATION,
    retention=LOG_RETENTION,
    compression="zip",
    backtrace=True,   # full stack on exceptions -> easier to fix
    diagnose=True,
    format="{time:YYYY-MM-DD HH:mm:ss} | {level: <8} | {name}:{function}:{line} - {message}",
)

__all__ = ["logger"]
