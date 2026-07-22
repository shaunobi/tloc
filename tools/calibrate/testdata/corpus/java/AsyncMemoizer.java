package example.net;

import java.util.Objects;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.function.Function;

public final class AsyncMemoizer<K, V> {
    private final ConcurrentHashMap<K, CompletableFuture<V>> pending = new ConcurrentHashMap<>();
    private final Function<K, CompletableFuture<V>> loader;

    public AsyncMemoizer(Function<K, CompletableFuture<V>> loader) {
        this.loader = Objects.requireNonNull(loader, "loader");
    }

    public CompletableFuture<V> get(K key) {
        Objects.requireNonNull(key, "key");
        return pending.computeIfAbsent(key, candidate -> {
            CompletableFuture<V> future = loader.apply(candidate);
            future.whenComplete((value, error) -> {
                if (error != null) {
                    pending.remove(candidate, future);
                }
            });
            return future;
        });
    }

    public boolean invalidate(K key) {
        return pending.remove(key) != null;
    }
}
