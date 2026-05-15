SELECT
  CASE
    WHEN COUNT(*) = 0 THEN 'missing'
    ELSE 'present'
  END AS endpoint_config_state
FROM file
WHERE path = '/Library/Application Support/Beacon/Endpoint/config.json';
