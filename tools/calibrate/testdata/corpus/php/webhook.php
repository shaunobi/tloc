<?php

declare(strict_types=1);

final class WebhookReceiver
{
    /** @param array<string, Closure(array): void> $handlers */
    public function __construct(
        private readonly string $secret,
        private array $handlers = [],
    ) {
        if ($secret === '') {
            throw new InvalidArgumentException('webhook secret is required');
        }
    }

    public function on(string $event, callable $handler): void
    {
        $this->handlers[$event] = Closure::fromCallable($handler);
    }

    public function receive(string $body, string $signature): void
    {
        $expected = hash_hmac('sha256', $body, $this->secret);
        if (!hash_equals($expected, strtolower($signature))) {
            throw new RuntimeException('signature verification failed');
        }

        $message = json_decode($body, true, flags: JSON_THROW_ON_ERROR);
        if (!is_array($message) || !is_string($message['type'] ?? null)) {
            throw new UnexpectedValueException('event type is missing');
        }

        $handler = $this->handlers[$message['type']] ?? null;
        if ($handler === null) {
            return;
        }

        $payload = $message['data'] ?? [];
        if (!is_array($payload)) {
            throw new UnexpectedValueException('event data must be an object');
        }
        $handler($payload);
    }
}
