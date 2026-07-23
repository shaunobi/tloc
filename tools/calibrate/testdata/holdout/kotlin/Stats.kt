package holdout.stats

import kotlin.math.sqrt

data class Summary(
    val count: Int,
    val minimum: Double,
    val maximum: Double,
    val mean: Double,
    val standardDeviation: Double,
)

fun summarize(values: List<Double>): Summary? {
    if (values.isEmpty()) return null
    var mean = 0.0
    var squaredDelta = 0.0
    var minimum = values.first()
    var maximum = values.first()

    values.forEachIndexed { index, value ->
        minimum = minOf(minimum, value)
        maximum = maxOf(maximum, value)
        val delta = value - mean
        mean += delta / (index + 1)
        squaredDelta += delta * (value - mean)
    }

    val variance = if (values.size > 1) squaredDelta / (values.size - 1) else 0.0
    return Summary(values.size, minimum, maximum, mean, sqrt(variance))
}

fun percentile(values: List<Double>, fraction: Double): Double {
    require(values.isNotEmpty() && fraction in 0.0..1.0)
    val sorted = values.sorted()
    val index = ((sorted.lastIndex) * fraction).toInt()
    return sorted[index]
}
