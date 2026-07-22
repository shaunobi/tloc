WITH signups AS (
    SELECT
        account_id,
        DATE_TRUNC('week', created_at)::date AS cohort_week
    FROM app.accounts
    WHERE created_at >= CURRENT_DATE - INTERVAL '180 days'
), activity AS (
    SELECT
        account_id,
        DATE_TRUNC('week', occurred_at)::date AS activity_week,
        COUNT(*) AS action_count
    FROM app.audit_events
    WHERE occurred_at >= CURRENT_DATE - INTERVAL '180 days'
    GROUP BY account_id, DATE_TRUNC('week', occurred_at)::date
), weekly AS (
    SELECT
        signups.cohort_week,
        activity.activity_week,
        ((activity.activity_week - signups.cohort_week) / 7)::integer AS week_number,
        COUNT(DISTINCT signups.account_id) AS active_accounts,
        SUM(activity.action_count) AS actions
    FROM signups
    JOIN activity USING (account_id)
    WHERE activity.activity_week >= signups.cohort_week
    GROUP BY signups.cohort_week, activity.activity_week
)
SELECT
    cohort_week,
    week_number,
    active_accounts,
    actions,
    ROUND(
        100.0 * active_accounts
        / NULLIF(FIRST_VALUE(active_accounts) OVER (
            PARTITION BY cohort_week ORDER BY week_number
        ), 0),
        1
    ) AS retention_percent,
    SUM(actions) OVER (
        PARTITION BY cohort_week ORDER BY week_number
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS cumulative_actions
FROM weekly
ORDER BY cohort_week, week_number;
