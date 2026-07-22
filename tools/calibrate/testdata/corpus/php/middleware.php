<?php

declare(strict_types=1);

final readonly class Request
{
    /** @param array<string, string> $headers */
    public function __construct(public string $path, public array $headers = [])
    {
    }
}

final readonly class Response
{
    public function __construct(public int $status, public string $body)
    {
    }
}

final class Pipeline
{
    /** @var list<Closure(Request, Closure): Response> */
    private array $middleware = [];

    public function pipe(callable $handler): self
    {
        $this->middleware[] = Closure::fromCallable($handler);
        return $this;
    }

    public function handle(Request $request, callable $destination): Response
    {
        $next = Closure::fromCallable($destination);
        foreach (array_reverse($this->middleware) as $middleware) {
            $downstream = $next;
            $next = static fn (Request $incoming): Response =>
                $middleware($incoming, $downstream);
        }
        return $next($request);
    }
}

function requireApiKey(Request $request, Closure $next): Response
{
    $provided = $request->headers['x-api-key'] ?? '';
    $expected = getenv('SERVICE_API_KEY') ?: '';
    if ($expected === '' || !hash_equals($expected, $provided)) {
        return new Response(401, json_encode(['error' => 'unauthorized'], JSON_THROW_ON_ERROR));
    }
    return $next($request);
}
