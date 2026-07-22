import asyncio
from collections.abc import Awaitable, Callable
from typing import TypeVar


T = TypeVar("T")


class AttemptsExhausted(RuntimeError):
    def __init__(self, attempts: int, cause: Exception) -> None:
        super().__init__(f"operation failed after {attempts} attempts")
        self.attempts = attempts
        self.__cause__ = cause


async def retry(
    operation: Callable[[], Awaitable[T]],
    *,
    attempts: int = 3,
    initial_delay: float = 0.05,
    retry_if: Callable[[Exception], bool] = lambda _: True,
) -> T:
    """Retry an asynchronous operation with capped exponential backoff."""
    if attempts < 1:
        raise ValueError("attempts must be positive")
    for number in range(1, attempts + 1):
        try:
            return await operation()
        except asyncio.CancelledError:
            raise
        except Exception as error:
            if number == attempts or not retry_if(error):
                raise AttemptsExhausted(number, error) from error
            await asyncio.sleep(min(initial_delay * 2 ** (number - 1), 2.0))
    raise AssertionError("retry loop exited unexpectedly")
