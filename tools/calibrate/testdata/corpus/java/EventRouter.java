package example.net;

import java.util.HashMap;
import java.util.Map;
import java.util.Objects;
import java.util.function.Consumer;

public final class EventRouter {
    private final Map<Class<?>, Consumer<Object>> handlers = new HashMap<>();

    public <T> void subscribe(Class<T> type, Consumer<? super T> handler) {
        Objects.requireNonNull(type, "type");
        Objects.requireNonNull(handler, "handler");
        handlers.put(type, event -> handler.accept(type.cast(event)));
    }

    public boolean publish(Object event) {
        Objects.requireNonNull(event, "event");
        Consumer<Object> handler = handlers.get(event.getClass());
        if (handler == null) {
            return false;
        }
        handler.accept(event);
        return true;
    }

    public int subscriptionCount() {
        return handlers.size();
    }
}
