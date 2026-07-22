require "json"
require "stringio"

module JsonLines
  module_function

  def each_record(io)
    return enum_for(__method__, io) unless block_given?

    io.each_line.with_index(1) do |line, number|
      next if line.strip.empty? || line.lstrip.start_with?("#")

      begin
        value = JSON.parse(line, symbolize_names: true)
        raise TypeError, "line #{number} must contain an object" unless value.is_a?(Hash)

        yield value.freeze
      rescue JSON::ParserError => error
        raise JSON::ParserError, "invalid JSON on line #{number}: #{error.message}"
      end
    end
  end

  def dump(records)
    output = StringIO.new
    records.each do |record|
      normalized = record.sort_by { |key, _| key.to_s }.to_h
      output.puts(JSON.generate(normalized))
    end
    output.string
  end
end
