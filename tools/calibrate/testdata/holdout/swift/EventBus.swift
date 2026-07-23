import Foundation

struct DomainEvent: Sendable {
    let name: String
    let attributes: [String: String]
    let occurredAt: Date
}

actor EventBus {
    typealias Handler = @Sendable (DomainEvent) async throws -> Void
    private var handlers: [String: [UUID: Handler]] = [:]

    func subscribe(to name: String, handler: @escaping Handler) -> UUID {
        let id = UUID()
        handlers[name, default: [:]][id] = handler
        return id
    }

    func unsubscribe(_ id: UUID, from name: String) {
        handlers[name]?[id] = nil
        if handlers[name]?.isEmpty == true {
            handlers[name] = nil
        }
    }

    func publish(_ event: DomainEvent) async -> [Error] {
        let selected = Array(handlers[event.name, default: [:]].values) +
            Array(handlers["*", default: [:]].values)
        var failures: [Error] = []
        for handler in selected {
            do { try await handler(event) } catch { failures.append(error) }
        }
        return failures
    }
}
