from collections import defaultdict
from dataclasses import dataclass
from decimal import Decimal


@dataclass(frozen=True, slots=True)
class StockItem:
    sku: str
    warehouse: str
    quantity: int
    unit_cost: Decimal

    def __post_init__(self) -> None:
        if self.quantity < 0:
            raise ValueError("quantity cannot be negative")
        if self.unit_cost < 0:
            raise ValueError("unit_cost cannot be negative")


def summarize(items: list[StockItem]) -> dict[str, dict[str, Decimal]]:
    """Return quantity and value totals grouped by warehouse."""
    result: dict[str, dict[str, Decimal]] = defaultdict(
        lambda: {"quantity": Decimal(0), "value": Decimal(0)}
    )
    for item in items:
        totals = result[item.warehouse]
        totals["quantity"] += item.quantity
        totals["value"] += item.unit_cost * item.quantity
    return dict(sorted(result.items()))
