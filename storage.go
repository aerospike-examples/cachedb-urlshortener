package main

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Storage interface {
	UrlStore() UrlStore
}

type storage struct {
	url *urlStore
}

func (s *storage) UrlStore() UrlStore {
	return s.url
}

func NewStorage(driver, src string) *storage {
	db, err := sql.Open(driver, src)
	if err != nil {
		panic(err)
	}

	if _, err = db.Exec(urlSchema); err != nil {
		panic(err)
	}

	return &storage{
		&urlStore{db},
	}
}

const urlSchema = `CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    hash varchar(255) NOT NULL,
    url varchar(255) NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS duplicateUrl ON urls(hash);`

type urlStore struct {
	*sql.DB
}

func (db *urlStore) Update(u *Url) error {
	err := db.QueryRow("INSERT INTO urls (hash, url) VALUES ($1, $2) RETURNING id", u.Hash, u.Url).Scan(&u.Id)
	if err != nil {
		return err
	}
	return err
}

func (db *urlStore) Get(hash string) (u Url, err error) {
	row := db.QueryRow("SELECT id, hash, url FROM urls WHERE hash = $1", hash)
	if err = row.Scan(&u.Id, &u.Hash, &u.Url); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return
	}
	return u, nil
}

func (db *urlStore) GetByUrl(url string) (u Url, err error) {
	row := db.QueryRow("SELECT id, hash, url FROM urls WHERE url = $1", url)
	if err = row.Scan(&u.Id, &u.Hash, &u.Url); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return
	}
	return u, nil
}

func (db *urlStore) Remove(hash string) error {
	_, err := db.Exec("DELETE FROM urls WHERE hash = $1", hash)
	return err
}
