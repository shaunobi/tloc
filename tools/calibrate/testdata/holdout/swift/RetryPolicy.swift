import Foundation

struct RetryPolicy: Sendable {
    let maximumAttempts: Int
    let initialDelay: Duration
    let maximumDelay: Duration
    let jitterFraction: Double

    init(maximumAttempts: Int, initialDelay: Duration, maximumDelay: Duration, jitterFraction: Double = 0.2) {
        precondition(maximumAttempts > 0)
        precondition(jitterFraction >= 0 && jitterFraction <= 1)
        self.maximumAttempts = maximumAttempts
        self.initialDelay = initialDelay
        self.maximumDelay = maximumDelay
        self.jitterFraction = jitterFraction
    }

    func delay(after attempt: Int, random: Double = .random(in: 0...1)) -> Duration? {
        guard attempt < maximumAttempts else { return nil }
        let exponent = min(attempt, 10)
        let baseSeconds = initialDelay.seconds * Double(1 << exponent)
        let cappedSeconds = min(baseSeconds, maximumDelay.seconds)
        let jitter = 1 + (random * 2 - 1) * jitterFraction
        return .seconds(cappedSeconds * jitter)
    }
}

private extension Duration {
    var seconds: Double {
        let parts = components
        return Double(parts.seconds) + Double(parts.attoseconds) / 1e18
    }
}
