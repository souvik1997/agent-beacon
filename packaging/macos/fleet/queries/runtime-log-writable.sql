SELECT
  CASE
    WHEN COUNT(*) = 0 THEN 'missing'
    ELSE 'present'
  END AS runtime_log_state
FROM file
WHERE path = '/var/log/beacon-agent/runtime.jsonl';
