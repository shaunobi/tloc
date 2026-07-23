package holdout.retry

import java.time.Duration
import java.time.Instant
import java.util.PriorityQueue

data class RetryJob(
    val key: String,
    val attempt: Int,
    val availableAt: Instant,
    val payload: String,
)

class RetryQueue {
    private val jobs = PriorityQueue<RetryJob>(compareBy(RetryJob::availableAt, RetryJob::key))

    @Synchronized
    fun schedule(key: String, payload: String, attempt: Int, now: Instant = Instant.now()) {
        require(attempt >= 0)
        val exponent = attempt.coerceAtMost(6)
        val delay = Duration.ofSeconds(1L shl exponent)
        jobs += RetryJob(key, attempt, now.plus(delay), payload)
    }

    @Synchronized
    fun takeReady(now: Instant = Instant.now(), limit: Int = 20): List<RetryJob> {
        require(limit > 0)
        return buildList {
            while (size < limit && jobs.peek()?.availableAt?.let { it <= now } == true) {
                add(jobs.remove())
            }
        }
    }

    @Synchronized
    fun size(): Int = jobs.size
}
