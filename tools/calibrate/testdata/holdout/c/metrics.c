#include <math.h>
#include <stddef.h>

typedef struct {
    double minimum;
    double maximum;
    double mean;
    double variance;
    size_t count;
} metric_summary;

metric_summary summarize_metrics(const double *values, size_t count) {
    metric_summary result = {0};
    if (values == NULL || count == 0) {
        return result;
    }

    double mean = 0.0;
    double squared_delta = 0.0;
    result.minimum = values[0];
    result.maximum = values[0];
    for (size_t index = 0; index < count; index++) {
        double value = values[index];
        if (value < result.minimum) result.minimum = value;
        if (value > result.maximum) result.maximum = value;

        double delta = value - mean;
        mean += delta / (double)(index + 1);
        squared_delta += delta * (value - mean);
    }
    result.mean = mean;
    result.variance = count > 1 ? squared_delta / (double)(count - 1) : 0.0;
    result.count = count;
    return result;
}

double metric_standard_deviation(metric_summary summary) {
    return sqrt(summary.variance);
}
