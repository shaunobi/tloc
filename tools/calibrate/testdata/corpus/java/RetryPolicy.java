package example.net;

import java.time.Duration;
import java.util.Objects;

public record RetryPolicy(int attempts, Duration initialDelay, double multiplier) {
    public RetryPolicy {
        if (attempts < 1) {
            throw new IllegalArgumentException("attempts must be positive");
        }
        Objects.requireNonNull(initialDelay, "initialDelay");
        if (initialDelay.isNegative() || multiplier < 1.0) {
            throw new IllegalArgumentException("invalid retry timing");
        }
    }

    public Duration delayForAttempt(int attempt) {
        if (attempt < 0 || attempt >= attempts) {
            throw new IndexOutOfBoundsException(attempt);
        }
        double scaled = initialDelay.toMillis() * Math.pow(multiplier, attempt);
        return Duration.ofMillis(Math.round(scaled));
    }
}
