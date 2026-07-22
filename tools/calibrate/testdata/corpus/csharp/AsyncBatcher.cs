using System.Runtime.CompilerServices;

namespace Example.Calibration;

public static class AsyncBatcher
{
    public static async IAsyncEnumerable<IReadOnlyList<T>> Batch<T>(
        IAsyncEnumerable<T> source,
        int size,
        [EnumeratorCancellation] CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(source);
        if (size <= 0)
        {
            throw new ArgumentOutOfRangeException(nameof(size), "Batch size must be positive.");
        }

        var pending = new List<T>(size);
        await foreach (var item in source.WithCancellation(cancellationToken))
        {
            pending.Add(item);
            if (pending.Count != size)
            {
                continue;
            }

            yield return pending.ToArray();
            pending.Clear();
        }

        if (pending.Count > 0)
        {
            yield return pending.ToArray();
        }
    }
}
