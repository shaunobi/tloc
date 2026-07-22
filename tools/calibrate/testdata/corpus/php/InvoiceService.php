<?php

declare(strict_types=1);

enum InvoiceState: string
{
    case Draft = 'draft';
    case Issued = 'issued';
    case Paid = 'paid';
    case Void = 'void';
}

final class Invoice
{
    public function __construct(
        public readonly string $number,
        public readonly int $amountCents,
        public InvoiceState $state = InvoiceState::Draft,
    ) {
        if ($number === '' || $amountCents < 0) {
            throw new InvalidArgumentException('invalid invoice data');
        }
    }
}

final class InvoiceService
{
    public function __construct(private array $invoices = [])
    {
    }

    public function issue(string $number): Invoice
    {
        $invoice = $this->find($number);
        if ($invoice->state !== InvoiceState::Draft) {
            throw new LogicException("invoice {$number} cannot be issued");
        }
        $invoice->state = InvoiceState::Issued;
        return $invoice;
    }

    public function markPaid(string $number): void
    {
        $invoice = $this->find($number);
        $invoice->state = match ($invoice->state) {
            InvoiceState::Issued => InvoiceState::Paid,
            InvoiceState::Paid => InvoiceState::Paid,
            default => throw new LogicException('only issued invoices can be paid'),
        };
    }

    private function find(string $number): Invoice
    {
        return $this->invoices[$number]
            ?? throw new OutOfBoundsException("invoice {$number} was not found");
    }
}
