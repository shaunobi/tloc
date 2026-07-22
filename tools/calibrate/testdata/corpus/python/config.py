from dataclasses import dataclass
from pathlib import Path
from typing import Any


@dataclass(frozen=True, slots=True)
class ServiceConfig:
    endpoint: str
    workers: int = 4
    labels: tuple[str, ...] = ()

    @classmethod
    def from_mapping(cls, raw: dict[str, Any]) -> "ServiceConfig":
        endpoint = str(raw.get("endpoint", "")).rstrip("/")
        if not endpoint.startswith(("http://", "https://")):
            raise ValueError("endpoint must be an HTTP URL")
        workers = int(raw.get("workers", 4))
        if not 1 <= workers <= 64:
            raise ValueError("workers must be between 1 and 64")
        labels = tuple(sorted({str(item).strip() for item in raw.get("labels", []) if item}))
        return cls(endpoint=endpoint, workers=workers, labels=labels)


def resolve_data_path(root: Path, relative: str) -> Path:
    candidate = (root / relative).resolve()
    if root.resolve() not in candidate.parents:
        raise ValueError(f"path escapes configured root: {relative!r}")
    return candidate
