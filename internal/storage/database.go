package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
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

	_, err = p.Exec(ctx, "CREATE TABLE IF NOT EXISTS link (id bigserial, short_link text, full_link text, user_id int, is_deleted bool default false)")
	if err != nil {
		return nil, err
	}
	_, err = p.Exec(ctx, "CREATE UNIQUE INDEX IF NOT EXISTS shortlink_indx ON link(short_link)")
	if err != nil {
		return nil, err
	}
	_, err = p.Exec(ctx, "CREATE UNIQUE INDEX IF NOT EXISTS full_link_indx ON link(full_link)")
	if err != nil {
		return nil, err
	}
	_, err = p.Exec(ctx, "CREATE INDEX IF NOT EXISTS full_link_indx ON link(user_id)")
	if err != nil {
		return nil, err
	}
	_, err = p.Exec(ctx, "CREATE TABLE IF NOT EXISTS auth_user (id bigserial)")
	if err != nil {
		return nil, err
	}
	return &Database{
		pool: p,
	}, nil

}

func (d *Database) Put(ctx context.Context, key string, val string, user int) error {

	query := `
		WITH inserted AS
			(INSERT INTO link (short_link, full_link, user_id)
			 VALUES ($1, $2, $3)
			 ON CONFLICT(full_link) DO NOTHING
			 RETURNING short_link)
		SELECT COALESCE (
			(SELECT short_link FROM inserted),
			(SELECT short_link FROM link WHERE full_link = $2)
		)`

	row := d.pool.QueryRow(ctx, query, key, val, user)

	var shortURL string
	if err := row.Scan(&shortURL); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return fmt.Errorf("%w", &KeyExistsError{Key: key})
		}
		return err
	}
	// if shortURL returned by DB differes from key, handle dublicate full_link
	if shortURL != key {
		return fmt.Errorf("%w", &ValueExistsError{Value: val, ExistingKey: shortURL})
	}
	return nil
}

func (d *Database) PutBatch(ctx context.Context, records ...URLRecord) error {
	batch := &pgx.Batch{}

	for _, rec := range records {
		batch.Queue("INSERT INTO link (short_link, full_link, user_id) VALUES ($1, $2, $3)", rec.ShortURL, rec.FullURL, rec.UserID)
	}
	br := d.pool.SendBatch(ctx, batch)
	return br.Close()
}

func (d *Database) DeleteBatch(ctx context.Context, records ...ToDelete) error {
	batch := &pgx.Batch{}

	for _, rec := range records {
		batch.Queue("UPDATE link set is_deleted = true WHERE short_link = $1 AND user_id = $2", rec.ShortURL, rec.UserID)
	}
	br := d.pool.SendBatch(ctx, batch)
	return br.Close()
}

func (d *Database) Get(ctx context.Context, key string) (string, error) {
	row := d.pool.QueryRow(ctx, "SELECT full_link, is_deleted FROM link WHERE short_link = $1", key)

	var URL string
	var is_deleted bool

	err := row.Scan(&URL, &is_deleted)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("%w", &KeyNotFoundError{Key: key})
	}
	if err != nil {
		return "", err
	}

	if is_deleted {
		return "", fmt.Errorf("%w", &RecordIsDeleted{Key: key})
	}

	return URL, nil
}

func (d *Database) CreateNewUser(ctx context.Context) (int, error) {
	row := d.pool.QueryRow(ctx, "INSERT INTO auth_user DEFAULT VALUES RETURNING id")

	var userID int
	err := row.Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (d *Database) GetUserURLS(ctx context.Context, userID int) ([]URLRecord, error) {
	rows, err := d.pool.Query(ctx, "SELECT short_link, full_link, user_id FROM link WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed collecting rows %w", err)
	}

	numbers, err := pgx.CollectRows(rows, pgx.RowToStructByName[URLRecord])
	if err != nil {
		return nil, fmt.Errorf("failed unpacking rows %w", err)
	}
	return numbers, nil
}

func (d *Database) Close() error {
	d.pool.Close()
	return nil
}
