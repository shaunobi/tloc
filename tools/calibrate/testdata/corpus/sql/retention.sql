WITH monthly_activity AS (
    SELECT
        user_id,
        DATE_TRUNC('month', occurred_at) AS activity_month,
        COUNT(*) AS event_count
    FROM analytics.events
    WHERE occurred_at >= CURRENT_DATE - INTERVAL '12 months'
    GROUP BY user_id, DATE_TRUNC('month', occurred_at)
), first_seen AS (
    SELECT user_id, MIN(activity_month) AS cohort_month
    FROM monthly_activity
    GROUP BY user_id
)
SELECT
    first_seen.cohort_month,
    activity.activity_month,
    COUNT(DISTINCT activity.user_id) AS active_users,
    SUM(activity.event_count) AS total_events
FROM monthly_activity AS activity
JOIN first_seen USING (user_id)
GROUP BY first_seen.cohort_month, activity.activity_month
ORDER BY first_seen.cohort_month, activity.activity_month;
