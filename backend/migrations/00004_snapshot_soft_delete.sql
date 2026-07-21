-- +goose Up
-- Soft-delete support for snapshots and debt_snapshots: a deleted row stays
-- in place (its holdings/entries stay attached, untouched) but is excluded
-- from every read path. The plain UNIQUE(user_id, snapshot_date) constraint
-- is replaced with a partial unique index scoped to non-deleted rows, so a
-- deleted date frees up for a fresh snapshot to reuse.
ALTER TABLE snapshots ADD COLUMN deleted_at timestamptz;
ALTER TABLE snapshots DROP CONSTRAINT snapshots_user_id_snapshot_date_key;
CREATE UNIQUE INDEX snapshots_user_date_active_uidx ON snapshots (user_id, snapshot_date) WHERE deleted_at IS NULL;

ALTER TABLE debt_snapshots ADD COLUMN deleted_at timestamptz;
ALTER TABLE debt_snapshots DROP CONSTRAINT debt_snapshots_user_id_snapshot_date_key;
CREATE UNIQUE INDEX debt_snapshots_user_date_active_uidx ON debt_snapshots (user_id, snapshot_date) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX debt_snapshots_user_date_active_uidx;
ALTER TABLE debt_snapshots ADD CONSTRAINT debt_snapshots_user_id_snapshot_date_key UNIQUE (user_id, snapshot_date);
ALTER TABLE debt_snapshots DROP COLUMN deleted_at;

DROP INDEX snapshots_user_date_active_uidx;
ALTER TABLE snapshots ADD CONSTRAINT snapshots_user_id_snapshot_date_key UNIQUE (user_id, snapshot_date);
ALTER TABLE snapshots DROP COLUMN deleted_at;
