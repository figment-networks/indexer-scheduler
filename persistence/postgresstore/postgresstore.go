package postgresstore

import (
	"context"
	"database/sql"
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
	rows, err := d.db.QueryContext(ctx, "SELECT id, run_id, network, chain_id, version, duration, kind, task_id, enabled FROM schedule")
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
		if err := rows.Scan(&rc.ID, &rc.RunID, &rc.Network, &rc.ChainID, &rc.Version, &rc.Duration, &rc.Kind, &rc.TaskID, &rc.Enabled); err != nil {
			return nil, err
		}
		rcs = append(rcs, rc)
	}

	return rcs, nil
}

func (d *Driver) MarkStatus(ctx context.Context, configID uuid.UUID, value bool) error {
	res, err := d.db.ExecContext(ctx, "UPDATE schedule SET enabled = $1 WHERE id = $2 ", value, configID)
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

func (d *Driver) MarkRunning(ctx context.Context, runID, configID uuid.UUID) error {
	res, err := d.db.ExecContext(ctx, "UPDATE schedule SET run_id = $1 WHERE id = $2 ", runID, configID)
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

func (d *Driver) AddConfig(ctx context.Context, rc structures.RunConfig) (err error) {

	var rID uuid.UUID
	var duration time.Duration

	row := d.db.QueryRowContext(ctx, "SELECT run_id, duration FROM schedule WHERE kind = $1 AND version = $2 AND network = $3 AND chain_id = $4 AND task_id = $5 ", rc.Kind, rc.Version, rc.Network, rc.ChainID, rc.TaskID)
	if err := row.Scan(&rID, &duration); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		res, err := d.db.ExecContext(ctx, "INSERT INTO schedule (run_id, network, version, chain_id, duration, kind, task_id, enabled) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", rc.RunID, rc.Network, rc.Version, rc.ChainID, rc.Duration, rc.Kind, rc.TaskID, rc.Enabled)
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
