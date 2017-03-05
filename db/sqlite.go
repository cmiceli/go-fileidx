package db

import (
	"database/sql"
	"github.com/golang/glog"
	_ "github.com/mattn/go-sqlite3"
)

type Database interface {
	Add(filename string, hash string) error
	CheckExistence(filename string) (bool, error)
}

type SQLiteDB struct {
	db *sql.DB
}

func (s *SQLiteDB) Setup() error {
	stmt, err := s.db.Prepare("CREATE TABLE files(filename VARCHAR(1024), hash VARCHAR(256), fullpath varchar(2048))")
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}

func (s *SQLiteDB) Add(filename string, hash string, full_path string) error {
	stmt, err := s.db.Prepare("INSERT INTO files(filename, hash, fullpath) values(?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(filename, hash, full_path)
	return err
}

func (s *SQLiteDB) CheckExistence(filename string, hash string) (bool, error) {
	stmt, err := s.db.Prepare("SELECT hash from files where filename = ?")
	if err != nil {
		return false, err
	}
	res, err := stmt.Query(filename)
	if err != nil {
		return false, err
	}
	var hashDb string
	for res.Next() {
		err = res.Scan(&hashDb)
		if err != nil {
			return false, err
		}
		glog.Infof("Hash: %s %s", hash, hashDb)
		if hash == hashDb {
			return true, nil
		}
	}
	return false, nil
}

func NewSQLiteDB(filename string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return &SQLiteDB{db: db}, nil
}
