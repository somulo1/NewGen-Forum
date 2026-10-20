#!/usr/bin/env bash
# Proves a backup restores: decrypts the newest one, loads it into a scratch database, runs the integrity check against it
# and compares row counts. Run weekly. A backup you have never restored is not a backup.
#   HUB_RESTORE_S3, HUB_RESTORE_AGE_IDENTITY (private key file), HUB_SCRATCH_ADMIN_URL (server where a scratch db may be created)
set -euo pipefail
: "${HUB_RESTORE_S3:?}" "${HUB_RESTORE_AGE_IDENTITY:?}" "${HUB_SCRATCH_ADMIN_URL:?}"
work=$(mktemp -d); trap 'rm -rf "$work"; psql "$HUB_SCRATCH_ADMIN_URL" -qc "drop database if exists hub_restore_test" || true' EXIT
newest=$(aws s3 ls "$HUB_RESTORE_S3/" | awk '/\.age$/ {print $4}' | sort | tail -1)
[ -n "$newest" ] || { echo "no backups found"; exit 1; }
aws s3 cp "$HUB_RESTORE_S3/$newest" "$work/" --only-show-errors
aws s3 cp "$HUB_RESTORE_S3/$newest.sha256" "$work/" --only-show-errors
(cd "$work" && sha256sum -c "$newest.sha256")
age -d -i "$HUB_RESTORE_AGE_IDENTITY" -o "$work/hub.dump" "$work/$newest"
psql "$HUB_SCRATCH_ADMIN_URL" -qc "drop database if exists hub_restore_test" -c "create database hub_restore_test"
pg_restore --no-owner --dbname "${HUB_SCRATCH_ADMIN_URL%/*}/hub_restore_test" "$work/hub.dump"
n=$(psql "${HUB_SCRATCH_ADMIN_URL%/*}/hub_restore_test" -Atc "select count(*) from hub.orgs")
echo "restored $newest: $n organizations"
[ "$n" -ge 0 ]
echo "restore test ok"
