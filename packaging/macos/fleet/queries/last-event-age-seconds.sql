SELECT
  CASE
    WHEN COUNT(*) = 0 THEN 'no_events'
    WHEN MAX(mtime) <= 0 THEN 'unknown'
    ELSE CAST((strftime('%s', 'now') - MAX(mtime)) AS TEXT)
  END AS last_event_age_seconds
FROM file
WHERE path = '/var/log/beacon-agent/runtime.jsonl'
  AND size > 0;
