package config

import (
	"database/sql"
	"fmt"
	"time"
)

type Postgres struct {
	Host     string `envconfig:"POSTGRES_HOST" required:"true"`
	Port     int    `envconfig:"POSTGRES_PORT" required:"true" default:"5432"`
	User     string `envconfig:"POSTGRES_USER" required:"true"`
	Password string `envconfig:"POSTGRES_PASSWORD" required:"true"`
	Dbname   string `envconfig:"POSTGRES_DB" required:"true" default:"postgres"`

	MaxConnectionLifetime          time.Duration `envconfig:"DB_MAX_CONN_LIFE_TIME" required:"true" default:"300s"`
	MaxOpenConnection              int           `envconfig:"DB_MAX_OPEN_CONNECTION" required:"true" default:"100"`
	MaxIdleConnection              int           `envconfig:"DB_MAX_IDLE_CONNECTION" required:"true" default:"10"`
	DBInitializationConnectTimeout int           `envconfig:"DB_INITIALIZATION_CONNECT_TIMEOUT" default:"2"`
}

func (p Postgres) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", p.Host, p.Port, p.User, p.Password, p.Dbname)
}

func (p Postgres) ConnectionStringWithTimeout() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s connect_timeout=%d",
		p.Host,
		p.Port,
		p.User,
		p.Password,
		p.Dbname,
		p.DBInitializationConnectTimeout)
}

func OpenDatabaseConnection(pg Postgres) (*sql.DB, error) {
	dbConn, err := sql.Open("postgres", pg.ConnectionString())
	if err != nil {
		return nil, err
	}

	dbConn.SetConnMaxLifetime(pg.MaxConnectionLifetime)
	dbConn.SetMaxOpenConns(pg.MaxOpenConnection)
	dbConn.SetMaxIdleConns(pg.MaxIdleConnection)

	err = dbConn.Ping()
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}
