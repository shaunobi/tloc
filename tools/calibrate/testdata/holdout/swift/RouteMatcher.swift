import Foundation

struct RouteMatch: Equatable {
    let handler: String
    let parameters: [String: String]
}

struct RouteMatcher {
    private struct Route {
        let segments: [String]
        let handler: String
    }

    private var routes: [Route] = []

    mutating func register(_ pattern: String, handler: String) {
        let segments = pattern.split(separator: "/").map(String.init)
        precondition(!segments.isEmpty)
        routes.append(Route(segments: segments, handler: handler))
    }

    func match(_ path: String) -> RouteMatch? {
        let incoming = path.split(separator: "/").map(String.init)
        for route in routes where route.segments.count == incoming.count {
            var parameters: [String: String] = [:]
            let accepted = zip(route.segments, incoming).allSatisfy { expected, actual in
                guard expected.hasPrefix(":") else { return expected == actual }
                parameters[String(expected.dropFirst())] = actual.removingPercentEncoding ?? actual
                return true
            }
            if accepted { return RouteMatch(handler: route.handler, parameters: parameters) }
        }
        return nil
    }
}
