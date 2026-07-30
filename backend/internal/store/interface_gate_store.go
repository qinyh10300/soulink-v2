package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type InterfaceGateWindow struct {
	GateKey   string
	StartsAt  time.Time
	EndsAt    time.Time
	Reason    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UpsertInterfaceGateWindowParams struct {
	GateKey   string
	StartsAt  time.Time
	EndsAt    time.Time
	Reason    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GetInterfaceGateWindow 查询指定 gate 的时间窗配置。
func (s *Store) GetInterfaceGateWindow(ctx context.Context, tx *sql.Tx, gateKey string) (InterfaceGateWindow, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
        SELECT gate_key, starts_at, ends_at, reason, created_at, updated_at
        FROM interface_gate_windows
        WHERE gate_key = $1
        LIMIT 1
    `, gateKey)

	var window InterfaceGateWindow
	if err := row.Scan(
		&window.GateKey,
		&window.StartsAt,
		&window.EndsAt,
		&window.Reason,
		&window.CreatedAt,
		&window.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return InterfaceGateWindow{}, ErrNotFound
		}
		return InterfaceGateWindow{}, fmt.Errorf("get interface gate window: %w", err)
	}
	return window, nil
}

// UpsertInterfaceGateWindow 保存或覆盖指定 gate 的关闭时间窗。
func (s *Store) UpsertInterfaceGateWindow(ctx context.Context, tx *sql.Tx, params UpsertInterfaceGateWindowParams) (InterfaceGateWindow, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
        INSERT INTO interface_gate_windows (
            gate_key, starts_at, ends_at, reason, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (gate_key)
        DO UPDATE SET
            starts_at = EXCLUDED.starts_at,
            ends_at = EXCLUDED.ends_at,
            reason = EXCLUDED.reason,
            updated_at = EXCLUDED.updated_at
        RETURNING gate_key, starts_at, ends_at, reason, created_at, updated_at
    `,
		params.GateKey,
		params.StartsAt,
		params.EndsAt,
		params.Reason,
		params.CreatedAt,
		params.UpdatedAt,
	)

	var window InterfaceGateWindow
	if err := row.Scan(
		&window.GateKey,
		&window.StartsAt,
		&window.EndsAt,
		&window.Reason,
		&window.CreatedAt,
		&window.UpdatedAt,
	); err != nil {
		return InterfaceGateWindow{}, fmt.Errorf("upsert interface gate window: %w", err)
	}
	return window, nil
}

// DeleteInterfaceGateWindow 删除指定 gate 的关闭时间窗。
func (s *Store) DeleteInterfaceGateWindow(ctx context.Context, tx *sql.Tx, gateKey string) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
        DELETE FROM interface_gate_windows
        WHERE gate_key = $1
    `, gateKey)
	if err != nil {
		return fmt.Errorf("delete interface gate window: %w", err)
	}
	return requireRowsAffected(result)
}
