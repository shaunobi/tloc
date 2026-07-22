<?php

declare(strict_types=1);

final readonly class Route
{
    public function __construct(
        public string $method,
        public string $pattern,
        public Closure $handler,
    ) {
        if ($pattern === '' || $pattern[0] !== '/') {
            throw new InvalidArgumentException('route pattern must start with /');
        }
    }
}

function matchRoute(array $routes, string $method, string $path): ?Route
{
    foreach ($routes as $route) {
        if ($route->method === strtoupper($method) && preg_match($route->pattern, $path) === 1) {
            return $route;
        }
    }
    return null;
}
