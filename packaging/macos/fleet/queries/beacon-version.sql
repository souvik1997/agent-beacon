SELECT
  CASE
    WHEN COUNT(*) = 0 THEN 'not_installed'
    ELSE 'installed'
  END AS beacon_install_state
FROM file
WHERE path = '/opt/beacon/bin/beacon';
