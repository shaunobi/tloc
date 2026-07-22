from collections.abc import Iterable, Iterator
from itertools import islice
from typing import TypeVar


T = TypeVar("T")


def batched(values: Iterable[T], size: int) -> Iterator[tuple[T, ...]]:
    """Yield fixed-size tuples without eagerly consuming the input."""
    if size < 1:
        raise ValueError("size must be at least one")
    iterator = iter(values)
    while batch := tuple(islice(iterator, size)):
        yield batch


def partition(values: Iterable[T], predicate) -> tuple[list[T], list[T]]:
    accepted: list[T] = []
    rejected: list[T] = []
    for value in values:
        (accepted if predicate(value) else rejected).append(value)
    return accepted, rejected


def unique_by(values: Iterable[T], key) -> Iterator[T]:
    seen: set[object] = set()
    for value in values:
        marker = key(value)
        if marker not in seen:
            seen.add(marker)
            yield value
