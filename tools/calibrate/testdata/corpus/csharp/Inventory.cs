namespace Example.Inventory;

public sealed record Item(string Sku, int Quantity, decimal UnitPrice);

public static class InventoryReport
{
    public static IReadOnlyDictionary<string, decimal> Totals(IEnumerable<Item> items)
    {
        ArgumentNullException.ThrowIfNull(items);
        return items
            .Where(item => item.Quantity > 0)
            .GroupBy(item => item.Sku, StringComparer.OrdinalIgnoreCase)
            .OrderBy(group => group.Key, StringComparer.OrdinalIgnoreCase)
            .ToDictionary(
                group => group.Key,
                group => group.Sum(item => item.Quantity * item.UnitPrice),
                StringComparer.OrdinalIgnoreCase);
    }
}
