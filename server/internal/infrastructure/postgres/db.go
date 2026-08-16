package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const pingTimeout = 5 * time.Second

// Open は DSN から接続プールを開く。呼び出し側で Close すること。
//
// ドライバには pgx（github.com/jackc/pgx/v5/stdlib）を使う。
// 上の blank import（_ "..."）で database/sql に "pgx" ドライバを登録し、
// sql.Open("pgx", ...) で標準の *sql.DB として扱える。
//
// sql.Open は接続の確立までは行わない（設定を持つだけ）。
// そのため PingContext で実際に DB へ疎通確認し、失敗時は Close してからエラーを返す。
// Ping にはタイムアウトを付け、DB 無応答時に起動が無限にブロックしないようにする。
func Open(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is empty")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
