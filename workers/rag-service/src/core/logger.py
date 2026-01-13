import logging
import logging.config
import os
from typing import Optional, Union


DEFAULT_FORMAT = "%(asctime)s [%(levelname)s] %(name)s: %(message)s"
DEFAULT_DATEFMT = "%Y-%m-%d %H:%M:%S"


def _coerce_level(level: Optional[Union[str, int]]) -> str:
    if isinstance(level, int):
        return logging.getLevelName(level)
    if isinstance(level, str):
        return level.upper()
    env_level = os.getenv("LOG_LEVEL", "INFO")
    return env_level.upper()


def configure_logging(
    level: Optional[Union[str, int]] = None, force: bool = False
) -> None:
    """
    全局配置日志。格式: "%(asctime)s [%(levelname)s] %(name)s: %(message)s"

    - 仅在尚未配置时生效，除非 force=True
    - 支持通过环境变量 LOG_LEVEL 指定日志级别
    """
    root_logger = logging.getLogger()
    if root_logger.handlers and not force:
        return

    effective_level = _coerce_level(level)

    config_dict = {
        "version": 1,
        "disable_existing_loggers": False,
        "formatters": {
            "standard": {
                "format": DEFAULT_FORMAT,
                "datefmt": DEFAULT_DATEFMT,
            }
        },
        "handlers": {
            "console": {
                "class": "logging.StreamHandler",
                "level": effective_level,
                "formatter": "standard",
                "stream": "ext://sys.stdout",
            }
        },
        "root": {
            "level": effective_level,
            "handlers": ["console"],
        },
        # 让 uvicorn/fastapi 的日志也用同样的格式，避免重复输出
        "loggers": {
            "uvicorn": {
                "level": effective_level,
                "handlers": ["console"],
                "propagate": False,
            },
            "uvicorn.error": {
                "level": effective_level,
                "handlers": ["console"],
                "propagate": False,
            },
            "uvicorn.access": {
                "level": effective_level,
                "handlers": ["console"],
                "propagate": False,
            },
        },
    }

    logging.config.dictConfig(config_dict)


def get_logger(name: str) -> logging.Logger:
    """获取命名 logger（如: "service"、"controller"、"repository" 等）。"""
    configure_logging()
    return logging.getLogger(name)


# 常用层级的便捷方法（可选使用）
def service_logger() -> logging.Logger:
    return get_logger("service")


def controller_logger() -> logging.Logger:
    return get_logger("controller")


def repository_logger() -> logging.Logger:
    return get_logger("repository")


def core_logger() -> logging.Logger:
    return get_logger("core")
