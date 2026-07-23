package holdout.config

sealed interface ConfigValue {
    data class Text(val value: String) : ConfigValue
    data class Number(val value: Double) : ConfigValue
    data class Flag(val value: Boolean) : ConfigValue
    data class Section(val entries: Map<String, ConfigValue>) : ConfigValue
}

fun mergeConfig(base: ConfigValue.Section, override: ConfigValue.Section): ConfigValue.Section {
    val keys = base.entries.keys + override.entries.keys
    val merged = keys.associateWith { key ->
        val left = base.entries[key]
        val right = override.entries[key]
        when {
            left is ConfigValue.Section && right is ConfigValue.Section -> mergeConfig(left, right)
            right != null -> right
            left != null -> left
            else -> error("unreachable key: $key")
        }
    }
    return ConfigValue.Section(merged)
}

fun ConfigValue.Section.lookup(path: String): ConfigValue? {
    var current: ConfigValue = this
    for (segment in path.split('.').filter(String::isNotBlank)) {
        current = (current as? ConfigValue.Section)?.entries?.get(segment) ?: return null
    }
    return current
}
