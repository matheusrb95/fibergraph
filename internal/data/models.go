package data

import (
	"database/sql"
	"errors"
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
