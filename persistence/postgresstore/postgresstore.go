package postgresstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/figment-networks/indexer-scheduler/persistence/params"
	"github.com/figment-networks/indexer-scheduler/structures"

	"github.com/google/uuid"
)

type Driver struct {
	db *sql.DB
}

func NewDriver(db *sql.DB) *Driver {
	return &Driver{
		db: db,
	}
}

func (d *Driver) GetConfigs(ctx context.Context) (rcs []structures.RunConfig, err error) {
	rows, err := d.db.QueryContext(ctx, "SELECT id, run_id, network, chain_id, version, duration, kind, task_id, enabled, status, config FROM schedule")
	switch {
	case err == sql.ErrNoRows:
		return nil, params.ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("query error: %w", err)
	default:
	}

	defer rows.Close()
	for rows.Next() {
		rc := structures.RunConfig{}

		configJSON := []byte{}
		if err := rows.Scan(&rc.ID, &rc.RunID, &rc.Network, &rc.ChainID, &rc.Version, &rc.Duration, &rc.Kind, &rc.TaskID, &rc.Enabled, &rc.Status, &configJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(configJSON, &rc.Config); err != nil {
			return nil, err
		}
		rcs = append(rcs, rc)
	}

	return rcs, nil
}

func (d *Driver) MarkRunning(ctx context.Context, runID, configID uuid.UUID) error {
	res, err := d.db.ExecContext(ctx, "UPDATE schedule SET run_id = $1, enabled = true, status = $2 WHERE id = $3 ", runID, structures.StateRunning, configID)
	if err != nil {
		return err
	}

	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i == 0 {
		return errors.New("no rows updated")
	}

	return nil
}

func (d *Driver) RemoveStatusAllEnabled(ctx context.Context) error {
	res, err := d.db.ExecContext(ctx, "UPDATE schedule SET status = $1 WHERE enabled = true", structures.StateAdded)
	if err != nil {
		return err
	}

	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i == 0 {
		return errors.New("no rows updated")
	}

	return nil
}

func (d *Driver) MarkStopped(ctx context.Context, id uuid.UUID) error {
	res, err := d.db.ExecContext(ctx, "UPDATE schedule SET enabled = false, status = $2 WHERE id = $1 ", id, structures.StateStopped)
	if err != nil {
		return err
	}

	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i == 0 {
		return errors.New("no rows updated")
	}

	return nil
}

func (d *Driver) MarkFinished(ctx context.Context, id uuid.UUID) error {
	res, err := d.db.ExecContext(ctx, "UPDATE schedule SET enabled = false, status = $2 WHERE id = $1", id, structures.StateFinished)
	if err != nil {
		return err
	}

	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i == 0 {
		return errors.New("no rows updated")
	}

	return nil
}

func (d *Driver) AddConfig(ctx context.Context, rc structures.RunConfig) (err error) {

	var rID uuid.UUID
	var duration time.Duration

	row := d.db.QueryRowContext(ctx, "SELECT run_id, duration FROM schedule WHERE kind = $1 AND version = $2 AND network = $3 AND chain_id = $4 AND task_id = $5", rc.Kind, rc.Version, rc.Network, rc.ChainID, rc.TaskID)
	if err := row.Scan(&rID, &duration); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		configJSON, err := json.Marshal(rc.Config)
		if err != nil {
			return err
		}

		res, err := d.db.ExecContext(ctx, "INSERT INTO schedule (run_id, network, version, chain_id, duration, kind, task_id, enabled, status, config) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)", rc.RunID, rc.Network, rc.Version, rc.ChainID, rc.Duration, rc.Kind, rc.TaskID, rc.Enabled, structures.StateAdded, configJSON)
		if err != nil {
			return err
		}

		i, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if i == 0 {
			return errors.New("No rows updated")
		}

		return nil
	}
	// Skip update for now
	/*
		if rc.RunID != rID {
			res, err := d.db.ExecContext(ctx, "UPDATE schedule SET duration = $1, run_id = $2 WHERE network = $3 AND version = $4 AND chain_id = $5 AND kind = $6 AND task_id = $7", rc.Duration, rc.RunID, rc.Network, rc.Version, rc.ChainID, rc.Kind, rc.TaskID)
			if err != nil {
				return err
			}
			i, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if i == 0 {
				return errors.New("No rows updated")
			}
			return nil
		} */

	return params.ErrAlreadyRegistred
}
