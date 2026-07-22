InvoiceLine = Struct.new(:description, :quantity, :unit_price, keyword_init: true) do
  def total
    raise ArgumentError, "quantity must be positive" unless quantity.positive?
    raise ArgumentError, "unit price cannot be negative" if unit_price.negative?

    quantity * unit_price
  end
end

class Invoice
  attr_reader :number, :lines

  def initialize(number, lines = [])
    @number = number.to_s.strip
    raise ArgumentError, "invoice number is required" if @number.empty?

    @lines = lines.dup
  end

  def subtotal
    lines.sum(&:total)
  end

  def totals_by_description
    lines
      .group_by(&:description)
      .transform_values { |items| items.sum(&:total) }
      .sort
      .to_h
  end

  def summary
    format("Invoice %<number>s: %<count>d lines, total %<total>.2f", number: number, count: lines.length, total: subtotal)
  end
end
