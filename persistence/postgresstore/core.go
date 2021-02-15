package postgresstore

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/figment-networks/indexer-scheduler/persistence/params"
	"github.com/figment-networks/indexer-scheduler/structures"
)

func (d *Driver) GetLatest(ctx context.Context, kind, network, chainID, taskID, version string) (lRec structures.LatestRecord, err error) {
	row := d.db.QueryRowContext(ctx, "SELECT hash, height, latesttime, nonce, retry, task_id FROM schedule_latest WHERE network = $1 AND chain_id = $2 AND version = $3 AND kind = $4 AND task_id = $5  ORDER BY time DESC LIMIT 1", network, chainID, version, kind, taskID)
	if row != nil {
		if err := row.Scan(&lRec.Hash, &lRec.Height, &lRec.Time, &lRec.Nonce, &lRec.RetryCount, &lRec.TaskID); err != nil {
			if err == sql.ErrNoRows {
				return lRec, structures.ErrDoesNotExists
			}

			return lRec, err
		}
	}
	return lRec, nil
}

func (d *Driver) SetLatest(ctx context.Context, kind, network, chainID, taskID, version string, lRec structures.LatestRecord) (err error) {
	_, err = d.db.ExecContext(ctx, "INSERT INTO schedule_latest (latesttime, network, chain_id, version, kind, task_id, hash, height, nonce, retry, error ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)",
		lRec.Time, network, chainID, version, kind, taskID, lRec.Hash, lRec.Height, lRec.Nonce, lRec.RetryCount, lRec.Error)
	return err
}

func (d *Driver) GetRuns(ctx context.Context, kind, network, taskID string, limit int) (lRec []structures.LatestRecord, err error) {

	q := "SELECT hash, height, latesttime, nonce, retry, error, task_id  FROM schedule_latest "

	var (
		args   []interface{}
		wherec []string
		i      = 1
	)

	if network != "" {
		wherec = append(wherec, ` network =  $`+strconv.Itoa(i))
		args = append(args, network)
		i++
	}
	if kind != "" {
		wherec = append(wherec, ` kind =  $`+strconv.Itoa(i))
		args = append(args, kind)
		i++
	}
	if taskID != "" {
		wherec = append(wherec, ` task_id =  $`+strconv.Itoa(i))
		args = append(args, taskID)
		i++
	}
	if len(args) > 0 {
		q += ` WHERE `
		q += strings.Join(wherec, " AND ")
	}

	q += ` ORDER BY time DESC LIMIT $` + strconv.Itoa(i)
	args = append(args, limit)

	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	switch {
	case err == sql.ErrNoRows:
		return nil, params.ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("query error: %w", err)
	default:
	}

	defer rows.Close()
	for rows.Next() {
		rc := structures.LatestRecord{}
		if err := rows.Scan(&rc.Hash, &rc.Height, &rc.Time, &rc.Nonce, &rc.RetryCount, &rc.Error, &rc.TaskID); err != nil {
			return nil, err
		}
		lRec = append(lRec, rc)
	}

	return lRec, nil
}
