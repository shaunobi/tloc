require "set"

class DependencyGraph
  def initialize
    @edges = Hash.new { |hash, key| hash[key] = Set.new }
  end

  def add(name, dependencies: [])
    @edges[name].merge(dependencies)
    dependencies.each { |dependency| @edges[dependency] }
    self
  end

  def installation_order
    temporary = Set.new
    permanent = Set.new
    ordered = []

    visit = lambda do |name|
      raise ArgumentError, "dependency cycle at #{name}" if temporary.include?(name)
      return if permanent.include?(name)

      temporary.add(name)
      @edges.fetch(name).sort.each { |dependency| visit.call(dependency) }
      temporary.delete(name)
      permanent.add(name)
      ordered << name
    end

    @edges.keys.sort.each { |name| visit.call(name) }
    ordered.freeze
  end
end
