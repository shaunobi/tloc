#include <charconv>
#include <map>
#include <optional>
#include <string>
#include <string_view>

namespace sample {

std::map<std::string, int> parse_limits(std::string_view text) {
    std::map<std::string, int> limits;
    while (!text.empty()) {
        const auto end = text.find('\n');
        auto line = text.substr(0, end);
        text = end == std::string_view::npos ? std::string_view{} : text.substr(end + 1);
        if (line.empty() || line.front() == '#') continue;

        const auto separator = line.find('=');
        if (separator == std::string_view::npos) continue;
        auto name = line.substr(0, separator);
        auto encoded = line.substr(separator + 1);
        int value = 0;
        const auto result = std::from_chars(encoded.data(), encoded.data() + encoded.size(), value);
        if (result.ec == std::errc{} && result.ptr == encoded.data() + encoded.size()) {
            limits.insert_or_assign(std::string{name}, value);
        }
    }
    return limits;
}

}  // namespace sample
