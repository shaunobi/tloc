module TextTools
  module_function

  def slug(value, separator: "-")
    normalized = value
      .unicode_normalize(:nfkd)
      .encode("ASCII", invalid: :replace, undef: :replace, replace: "")
      .downcase
      .gsub(/[^a-z0-9]+/, separator)
      .gsub(/#{Regexp.escape(separator)}{2,}/, separator)

    normalized.delete_prefix(separator).delete_suffix(separator)
  end

  def unique_slugs(values)
    counts = Hash.new(0)
    values.map do |value|
      base = slug(value)
      counts[base] += 1
      counts[base] == 1 ? base : "#{base}-#{counts[base]}"
    end
  end
end
