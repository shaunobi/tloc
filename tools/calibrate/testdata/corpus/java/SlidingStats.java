package example.net;

import java.util.ArrayDeque;
import java.util.Deque;

public final class SlidingStats {
    public record Snapshot(int count, double minimum, double maximum, double average) {}

    private final int capacity;
    private final Deque<Double> values = new ArrayDeque<>();

    public SlidingStats(int capacity) {
        if (capacity < 1) throw new IllegalArgumentException("capacity must be positive");
        this.capacity = capacity;
    }

    public synchronized void add(double value) {
        if (!Double.isFinite(value)) throw new IllegalArgumentException("value must be finite");
        values.addLast(value);
        if (values.size() > capacity) values.removeFirst();
    }

    public synchronized Snapshot snapshot() {
        if (values.isEmpty()) return new Snapshot(0, 0, 0, 0);
        double min = Double.POSITIVE_INFINITY;
        double max = Double.NEGATIVE_INFINITY;
        double sum = 0;
        for (double value : values) {
            min = Math.min(min, value);
            max = Math.max(max, value);
            sum += value;
        }
        return new Snapshot(values.size(), min, max, sum / values.size());
    }
}
