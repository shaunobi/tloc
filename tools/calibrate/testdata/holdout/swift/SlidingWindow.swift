import Foundation

struct SlidingWindow {
    private var values: [(timestamp: Date, value: Double)] = []
    let duration: TimeInterval

    init(duration: TimeInterval) {
        precondition(duration > 0)
        self.duration = duration
    }

    mutating func append(_ value: Double, at timestamp: Date = Date()) {
        values.append((timestamp, value))
        evict(before: timestamp.addingTimeInterval(-duration))
    }

    mutating func summary(now: Date = Date()) -> (count: Int, mean: Double, maximum: Double?) {
        evict(before: now.addingTimeInterval(-duration))
        guard !values.isEmpty else { return (0, 0, nil) }
        let total = values.reduce(0) { $0 + $1.value }
        return (values.count, total / Double(values.count), values.map(\.value).max())
    }

    private mutating func evict(before cutoff: Date) {
        if let firstCurrent = values.firstIndex(where: { $0.timestamp >= cutoff }) {
            values.removeFirst(firstCurrent)
        } else {
            values.removeAll(keepingCapacity: true)
        }
    }
}
