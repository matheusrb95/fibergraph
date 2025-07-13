package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrRecordNotFound = errors.New("record not found")

type Models struct {
	Connection ConnectionModel
}

func NewModels(db *sql.DB) *Models {
	return &Models{
		Connection: ConnectionModel{DB: db},
	}
}

func setSchema(ctx context.Context, tx *sql.Tx, tenantID string) error {
	query := fmt.Sprintf("USE `fkcp_db_ospmanager-%s`;", tenantID)

	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
