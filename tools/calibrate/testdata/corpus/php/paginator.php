<?php

declare(strict_types=1);

/** @template T */
final readonly class Page implements IteratorAggregate, Countable
{
    /** @param list<T> $items */
    public function __construct(
        public array $items,
        public int $number,
        public int $pageSize,
        public int $total,
    ) {
        if ($number < 1 || $pageSize < 1 || $total < count($items)) {
            throw new InvalidArgumentException('invalid pagination metadata');
        }
    }

    public function count(): int
    {
        return count($this->items);
    }

    public function getIterator(): Traversable
    {
        yield from $this->items;
    }

    public function hasNext(): bool
    {
        return $this->number * $this->pageSize < $this->total;
    }

    public function map(callable $transform): self
    {
        return new self(
            array_values(array_map($transform, $this->items)),
            $this->number,
            $this->pageSize,
            $this->total,
        );
    }
}

/** @return Page<array{id: int, name: string}> */
function paginateUsers(array $rows, int $page, int $size, int $total): Page
{
    $users = array_map(
        static fn (array $row): array => ['id' => (int) $row['id'], 'name' => trim($row['name'])],
        $rows,
    );
    return new Page($users, $page, $size, $total);
}
