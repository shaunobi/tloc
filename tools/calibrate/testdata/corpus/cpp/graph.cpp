#include <queue>
#include <stdexcept>
#include <vector>

namespace sample {

class graph {
public:
    explicit graph(std::size_t vertices) : edges_(vertices) {}

    void connect(std::size_t from, std::size_t to) {
        if (from >= edges_.size() || to >= edges_.size()) {
            throw std::out_of_range("vertex index");
        }
        edges_[from].push_back(to);
    }

    std::vector<std::size_t> breadth_first(std::size_t start) const {
        if (start >= edges_.size()) throw std::out_of_range("start vertex");
        std::vector<std::size_t> order;
        std::vector<bool> seen(edges_.size());
        std::queue<std::size_t> pending;
        pending.push(start);
        seen[start] = true;
        while (!pending.empty()) {
            const auto vertex = pending.front();
            pending.pop();
            order.push_back(vertex);
            for (auto neighbor : edges_[vertex]) {
                if (!seen[neighbor]) {
                    seen[neighbor] = true;
                    pending.push(neighbor);
                }
            }
        }
        return order;
    }

private:
    std::vector<std::vector<std::size_t>> edges_;
};

}  // namespace sample
