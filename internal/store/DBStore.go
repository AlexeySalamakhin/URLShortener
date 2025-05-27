package store

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
) // go postgres driver

type DBStore struct {
	DBReady bool
}

func NewDBStore(ConnectionString string) (*DBStore, error) {
	db, err := sql.Open("pgx", ConnectionString)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var dbStore DBStore
	err = db.Ping()
	if err == nil {
		dbStore.DBReady = true
	}

	return &dbStore, nil
}

func (dbStore *DBStore) Save(originalURL string, shortURL string) error {
	return nil
}
func (dbStore *DBStore) Get(shortURL string) (found bool, originalURL string) {
	return true, ""
}
func (dbStore *DBStore) Ready() bool {
	return dbStore.DBReady
}
