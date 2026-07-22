class CircuitBreaker
  class OpenError < StandardError; end

  def initialize(failure_limit: 3, cooldown: 30, clock: -> { Process.clock_gettime(Process::CLOCK_MONOTONIC) })
    raise ArgumentError, "failure_limit must be positive" unless failure_limit.positive?

    @failure_limit = failure_limit
    @cooldown = cooldown
    @clock = clock
    @failures = 0
    @opened_at = nil
  end

  def call
    if @opened_at && @clock.call - @opened_at < @cooldown
      raise OpenError, "circuit remains open"
    end

    result = yield
    @failures = 0
    @opened_at = nil
    result
  rescue OpenError
    raise
  rescue StandardError
    @failures += 1
    @opened_at = @clock.call if @failures >= @failure_limit
    raise
  end
end
