namespace Example.Calibration;

public sealed class MetricsWindow
{
    private readonly Queue<double> _values = new();
    private readonly object _gate = new();

    public MetricsWindow(int capacity)
    {
        Capacity = capacity > 0
            ? capacity
            : throw new ArgumentOutOfRangeException(nameof(capacity));
    }

    public int Capacity { get; }

    public void Record(double value)
    {
        if (!double.IsFinite(value))
        {
            throw new ArgumentException("Metric must be finite.", nameof(value));
        }

        lock (_gate)
        {
            _values.Enqueue(value);
            while (_values.Count > Capacity)
            {
                _values.Dequeue();
            }
        }
    }

    public (int Count, double Average) Snapshot()
    {
        lock (_gate)
        {
            return (_values.Count, _values.Count == 0 ? 0 : _values.Average());
        }
    }
}
