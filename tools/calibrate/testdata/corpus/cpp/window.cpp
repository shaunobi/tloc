#include <algorithm>
#include <stdexcept>
#include <vector>

namespace sample {

template <typename T>
std::vector<T> moving_average(const std::vector<T>& values, std::size_t width) {
    if (width == 0 || width > values.size()) {
        throw std::invalid_argument("width is outside the input range");
    }
    std::vector<T> result;
    result.reserve(values.size() - width + 1);
    T sum{};
    for (std::size_t index = 0; index < values.size(); ++index) {
        sum += values[index];
        if (index >= width) {
            sum -= values[index - width];
        }
        if (index + 1 >= width) {
            result.push_back(sum / static_cast<T>(width));
        }
    }
    return result;
}

template std::vector<double> moving_average(const std::vector<double>&, std::size_t);
}  // namespace sample
