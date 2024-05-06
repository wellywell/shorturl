package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
}

func NewDatabase(connString string) (*Database, error) {

	ctx := context.Background()
	p, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	_, err = p.Exec(ctx, "CREATE TABLE IF NOT EXISTS link (id bigserial, short_link text, full_link text)")
	if err != nil {
		return nil, err
	}
	_, err = p.Exec(ctx, "CREATE UNIQUE INDEX IF NOT EXISTS shortlink_indx ON link(short_link)")
	if err != nil {
		return nil, err
	}
	return &Database{
		pool: p,
	}, nil

}

func (d *Database) Put(ctx context.Context, key string, val string) error {
	_, err := d.pool.Exec(ctx, "INSERT INTO link (short_link, full_link) VALUES ($1, $2)", key, val)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == UniqueViolationErr {
			return &KeyExistsError{Key: key}
		}
	}
	return err
}

func (d *Database) PutBatch(ctx context.Context, records ...KeyValue) error {
	batch := &pgx.Batch{}

	for _, rec := range records {
		batch.Queue("INSERT INTO link (short_link, full_link) VALUES ($1, $2)", rec.Key, rec.Value)
	}
	br := d.pool.SendBatch(ctx, batch)
	return br.Close()
}

func (d *Database) Get(ctx context.Context, key string) (string, error) {
	row := d.pool.QueryRow(ctx, "SELECT full_link FROM link WHERE short_link = $1", key)

	var URL string
	err := row.Scan(&URL)

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return "", &KeyNotFoundError{Key: key}
	}
	if err != nil {
		return "", err
	}
	return URL, nil
}

func (d *Database) Close() error {
	d.pool.Close()
	return nil
}
