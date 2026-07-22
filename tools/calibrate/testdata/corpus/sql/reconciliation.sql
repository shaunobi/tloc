WITH expected AS (
    SELECT
        payment_id,
        invoice_id,
        amount_cents,
        currency,
        settled_at::date AS settlement_date
    FROM billing.payments
    WHERE settled_at >= :window_start
      AND settled_at < :window_end
      AND status = 'settled'
), received AS (
    SELECT
        processor_reference AS payment_id,
        SUM(net_amount_cents) AS amount_cents,
        MIN(currency) AS currency,
        MIN(deposit_date) AS deposit_date
    FROM imports.processor_deposits
    WHERE deposit_date BETWEEN :window_start::date AND :window_end::date
    GROUP BY processor_reference
), differences AS (
    SELECT
        expected.payment_id,
        expected.invoice_id,
        expected.amount_cents AS expected_cents,
        received.amount_cents AS received_cents,
        COALESCE(received.amount_cents, 0) - expected.amount_cents AS variance_cents,
        received.deposit_date
    FROM expected
    LEFT JOIN received USING (payment_id, currency)
)
INSERT INTO billing.reconciliation_issue (
    payment_id, invoice_id, expected_cents, received_cents, variance_cents, detected_at
)
SELECT payment_id, invoice_id, expected_cents, received_cents, variance_cents, CURRENT_TIMESTAMP
FROM differences
WHERE received_cents IS NULL OR variance_cents <> 0
ON CONFLICT (payment_id) DO UPDATE SET
    received_cents = EXCLUDED.received_cents,
    variance_cents = EXCLUDED.variance_cents,
    detected_at = EXCLUDED.detected_at;
