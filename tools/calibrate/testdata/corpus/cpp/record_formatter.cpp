#include <iomanip>
#include <sstream>
#include <string>
#include <type_traits>
#include <variant>
#include <vector>

namespace sample {

using field = std::variant<std::string, long, double, bool>;

std::string format_record(const std::vector<field>& fields) {
    std::ostringstream output;
    output << '{';
    for (std::size_t index = 0; index < fields.size(); ++index) {
        if (index != 0) output << ", ";
        std::visit([&output](const auto& value) {
            using value_type = std::decay_t<decltype(value)>;
            if constexpr (std::is_same_v<value_type, std::string>) {
                output << std::quoted(value);
            } else if constexpr (std::is_same_v<value_type, bool>) {
                output << (value ? "true" : "false");
            } else {
                output << value;
            }
        }, fields[index]);
    }
    output << '}';
    return output.str();
}

}  // namespace sample
