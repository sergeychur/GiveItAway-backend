package database

import (
	_ "github.com/lib/pq"
	"gopkg.in/jackc/pgx.v2"

	"time"
)

const (
	OK           = 0
	FOUND        = 0
	CREATED      = 1
	DB_ERROR     = 2
	EMPTY_RESULT = 3
	FORBIDDEN    = 4
	CONFLICT     = 5
	WRONG_INPUT  = 6
	TOO_MUCH_TIMES = 7
)

type DB struct {
	db           *pgx.ConnPool
	user         string
	password     string
	databaseName string
	host         string
	port         uint16

	AntiFloodAdMap map[int]map[int][]time.Time
}

func NewDB(user string, password string, dataBaseName string,
	host string, port uint16) *DB {
	db := new(DB)
	db.user = user
	db.databaseName = dataBaseName
	db.password = password
	db.host = host
	db.port = port

	db.AntiFloodAdMap = make(map[int]map[int][]time.Time)
	return db
}

func (db *DB) Start() error {
	conf := pgx.ConnConfig{
		Host:     db.host,
		Port:     db.port,
		User:     db.user,
		Password: db.password,
		Database: db.databaseName,
	}
	poolConf := pgx.ConnPoolConfig{
		ConnConfig:     conf,
		MaxConnections: 80,
		AcquireTimeout: time.Duration(1 * time.Second),
	}
	dataBase, err := pgx.NewConnPool(poolConf)
	if err != nil {
		return err
	}

	db.db = dataBase
	return nil
}

func (db *DB) Close() {
	db.db.Close()
}

func (db *DB) StartTransaction() (*pgx.Tx, error) {
	return db.db.Begin()
}
