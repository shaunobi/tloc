WITH ordered_events AS (
    SELECT
        user_id,
        occurred_at,
        event_name,
        LAG(occurred_at) OVER (
            PARTITION BY user_id ORDER BY occurred_at
        ) AS previous_at
    FROM analytics.events
    WHERE occurred_at >= :report_date::date
      AND occurred_at < :report_date::date + INTERVAL '1 day'
), boundaries AS (
    SELECT
        *,
        CASE
            WHEN previous_at IS NULL
              OR occurred_at - previous_at > INTERVAL '30 minutes'
            THEN 1 ELSE 0
        END AS starts_session
    FROM ordered_events
), numbered AS (
    SELECT
        *,
        SUM(starts_session) OVER (
            PARTITION BY user_id ORDER BY occurred_at
            ROWS UNBOUNDED PRECEDING
        ) AS session_number
    FROM boundaries
)
INSERT INTO analytics.daily_sessions (
    report_date, user_id, session_number, started_at, ended_at, event_count, converted
)
SELECT
    :report_date::date,
    user_id,
    session_number,
    MIN(occurred_at),
    MAX(occurred_at),
    COUNT(*),
    BOOL_OR(event_name = 'purchase_completed')
FROM numbered
GROUP BY user_id, session_number
ON CONFLICT (report_date, user_id, session_number) DO UPDATE SET
    ended_at = EXCLUDED.ended_at,
    event_count = EXCLUDED.event_count,
    converted = EXCLUDED.converted;
