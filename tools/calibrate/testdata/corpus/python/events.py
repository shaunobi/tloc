from dataclasses import dataclass
from datetime import datetime, timezone


@dataclass(frozen=True)
class Event:
    name: str
    occurred_at: datetime
    labels: tuple[str, ...] = ()


def parse_event(record: dict[str, object]) -> Event:
    """Validate a JSON-like record and normalize its labels."""
    name = str(record.get("name", "")).strip()
    if not name:
        raise ValueError("event name is required")
    raw_time = str(record["occurred_at"]).replace("Z", "+00:00")
    occurred_at = datetime.fromisoformat(raw_time).astimezone(timezone.utc)
    labels = tuple(sorted({str(value).lower() for value in record.get("labels", [])}))
    return Event(name=name, occurred_at=occurred_at, labels=labels)


def group_by_day(events: list[Event]) -> dict[str, list[Event]]:
    result: dict[str, list[Event]] = {}
    for event in events:
        result.setdefault(event.occurred_at.date().isoformat(), []).append(event)
    return result
