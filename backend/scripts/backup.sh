#!/usr/bin/env bash
# Daily database backup: a compressed dump, checked to be readable, copied to storage that cannot be altered or deleted
# (S3 with Object Lock), and a heartbeat file for monitoring. Run from cron or a systemd timer.
#   HUB_BACKUP_DB_URL   connection string for a read-only backup role (not the app or owner role)
#   HUB_BACKUP_DIR      local staging directory
#   HUB_BACKUP_S3       s3://bucket/prefix (bucket created with Object Lock in compliance mode, retention >= 35 days)
#   HUB_BACKUP_AGE_KEY  age public key; backups are encrypted before they leave the machine
#   HUB_METRICS_DIR     node_exporter textfile directory for the heartbeat (optional)
set -euo pipefail
: "${HUB_BACKUP_DB_URL:?}" "${HUB_BACKUP_DIR:?}" "${HUB_BACKUP_S3:?}" "${HUB_BACKUP_AGE_KEY:?}"
stamp=$(date -u +%Y%m%dT%H%M%SZ)
f="$HUB_BACKUP_DIR/hub-$stamp.dump"
mkdir -p "$HUB_BACKUP_DIR"
pg_dump --format=custom --compress=6 --no-owner --file "$f" "$HUB_BACKUP_DB_URL"
pg_restore --list "$f" > /dev/null                 # the archive must be readable
age -r "$HUB_BACKUP_AGE_KEY" -o "$f.age" "$f" && rm -f "$f"
sha256sum "$f.age" > "$f.age.sha256"
aws s3 cp "$f.age" "$HUB_BACKUP_S3/" --only-show-errors
aws s3 cp "$f.age.sha256" "$HUB_BACKUP_S3/" --only-show-errors
find "$HUB_BACKUP_DIR" -name 'hub-*.age*' -mtime +3 -delete   # keep three days locally; the bucket holds the rest
if [ -n "${HUB_METRICS_DIR:-}" ]; then
  printf '# TYPE hub_backup_last_success_timestamp_seconds gauge\nhub_backup_last_success_timestamp_seconds %s\n' "$(date +%s)" > "$HUB_METRICS_DIR/hub_backup.prom.$$" && mv "$HUB_METRICS_DIR/hub_backup.prom.$$" "$HUB_METRICS_DIR/hub_backup.prom"
fi
echo "backup ok: $f.age"
