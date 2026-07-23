import Foundation

actor AsyncCache<Key: Hashable, Value> {
    private struct Entry {
        let value: Value
        let expiresAt: Date
    }

    private var entries: [Key: Entry] = [:]
    private let lifetime: TimeInterval

    init(lifetime: TimeInterval) {
        precondition(lifetime > 0)
        self.lifetime = lifetime
    }

    func value(for key: Key, now: Date = Date()) -> Value? {
        guard let entry = entries[key] else { return nil }
        guard entry.expiresAt > now else {
            entries.removeValue(forKey: key)
            return nil
        }
        return entry.value
    }

    func insert(_ value: Value, for key: Key, now: Date = Date()) {
        entries[key] = Entry(value: value, expiresAt: now.addingTimeInterval(lifetime))
    }

    func removeExpired(now: Date = Date()) -> Int {
        let expired = entries.compactMap { key, entry in entry.expiresAt <= now ? key : nil }
        expired.forEach { entries.removeValue(forKey: $0) }
        return expired.count
    }
}
