package storage

type URLRecord struct {
	ShortURL string `db:"short_link"`
	FullURL  string `db:"full_link"`
	UserID   int    `db:"user_id"`
}
