package holdout.batch

import java.time.Duration
import java.time.Instant

data class BatchItem(val id: String, val payload: ByteArray, val receivedAt: Instant)

class BatchWindow(
    private val maximumItems: Int,
    private val maximumAge: Duration,
    private val clock: () -> Instant = Instant::now,
) {
    private val pending = ArrayDeque<BatchItem>()

    fun add(item: BatchItem): List<BatchItem>? {
        require(item.id.isNotBlank()) { "item id cannot be blank" }
        pending.addLast(item)
        return if (shouldFlush()) drain() else null
    }

    fun flushExpired(): List<BatchItem> = if (shouldFlushByAge()) drain() else emptyList()

    private fun shouldFlush(): Boolean = pending.size >= maximumItems || shouldFlushByAge()

    private fun shouldFlushByAge(): Boolean {
        val oldest = pending.firstOrNull() ?: return false
        return Duration.between(oldest.receivedAt, clock()) >= maximumAge
    }

    private fun drain(): List<BatchItem> = buildList(pending.size) {
        while (pending.isNotEmpty()) add(pending.removeFirst())
    }
}
