#include <condition_variable>
#include <cstddef>
#include <mutex>
#include <optional>
#include <queue>
#include <stdexcept>
#include <utility>

namespace sample {

template <typename T>
class bounded_queue {
public:
    explicit bounded_queue(std::size_t capacity) : capacity_(capacity) {
        if (capacity == 0) throw std::invalid_argument("capacity must be positive");
    }

    bool push(T value) {
        std::unique_lock lock(mutex_);
        ready_.wait(lock, [this] { return closed_ || values_.size() < capacity_; });
        if (closed_) return false;
        values_.push(std::move(value));
        ready_.notify_all();
        return true;
    }

    std::optional<T> pop() {
        std::unique_lock lock(mutex_);
        ready_.wait(lock, [this] { return closed_ || !values_.empty(); });
        if (values_.empty()) return std::nullopt;
        T value = std::move(values_.front());
        values_.pop();
        ready_.notify_all();
        return value;
    }

    void close() {
        std::lock_guard lock(mutex_);
        closed_ = true;
        ready_.notify_all();
    }

private:
    std::size_t capacity_;
    std::queue<T> values_;
    std::mutex mutex_;
    std::condition_variable ready_;
    bool closed_ = false;
};

template class bounded_queue<int>;
}  // namespace sample
