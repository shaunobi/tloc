package holdout.events

import java.util.concurrent.ConcurrentHashMap

data class Event(val topic: String, val attributes: Map<String, String>, val body: String)
typealias EventHandler = (Event) -> Unit

class EventRouter {
    private val handlers = ConcurrentHashMap<String, MutableList<EventHandler>>()

    fun subscribe(topic: String, handler: EventHandler): AutoCloseable {
        require(topic.isNotBlank())
        val topicHandlers = handlers.computeIfAbsent(topic) { mutableListOf() }
        synchronized(topicHandlers) { topicHandlers += handler }
        return AutoCloseable {
            synchronized(topicHandlers) {
                topicHandlers.remove(handler)
                if (topicHandlers.isEmpty()) handlers.remove(topic, topicHandlers)
            }
        }
    }

    fun publish(event: Event): Int {
        val selected = buildList {
            handlers[event.topic]?.let { synchronized(it) { addAll(it) } }
            handlers["*"]?.let { synchronized(it) { addAll(it) } }
        }
        selected.forEach { handler -> handler(event) }
        return selected.size
    }
}
