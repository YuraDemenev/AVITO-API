package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Константы с именами таблиц
const (
	adminsTable      = "admins"
	usersTable       = "users"
	tagsTable        = "tags"
	featuresTable    = "features"
	bannersTable     = "banners"
	bannersTagsTable = "banners_tags"
)

// Структура для передачи данных необходимых
// Чтобы подключиться к DB
type DBConfig struct {
	Host     string
	Port     string
	UserName string
	Password string
	DBName   string
	SSLMode  string
}

// Функция подключения к Postgres DB
func NewPostgresDB(cfg DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.UserName, cfg.DBName, cfg.Password, cfg.SSLMode))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
