package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open は DSN から接続プールを開く。呼び出し側で Close すること。
//
// ドライバには pgx（github.com/jackc/pgx/v5/stdlib）を使う。
// 上の blank import（_ "..."）で database/sql に "pgx" ドライバを登録し、
// sql.Open("pgx", ...) で標準の *sql.DB として扱える。
//
// sql.Open は接続の確立までは行わない（設定を持つだけ）。
// そのため Ping で実際に DB へ疎通確認し、失敗時は Close してからエラーを返す。
func Open(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is empty")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
