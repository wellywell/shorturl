package storage

// URLRecord информация о ссылке
type URLRecord struct {
	ShortURL  string `db:"short_link"`
	FullURL   string `db:"full_link"`
	UserID    int    `db:"user_id"`
	IsDeleted bool   `db:"is_deleted"`
}

// ToDelete структура для создание тасок на удаление ссылки
type ToDelete struct {
	ShortURL string
	UserID   int
}
