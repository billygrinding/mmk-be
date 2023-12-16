package postgres

import (
	"context"
	"database/sql"
	"github.com/billygrinding/mmk-be/pkg/dbresolver"
	"time"

	// This is imported for migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/stretchr/testify/suite"
)

const (
	postgres   = "postgres"
	DsnDefault = "user=user password=password dbname=mmk_be host=localhost port=5432 sslmode=disable"
)

// Suite struct for MySQL Suite
type Suite struct {
	suite.Suite
	DSN                     string
	DBConn                  *dbresolver.DB
	Migration               *migration
	MigrationLocationFolder string
	DBName                  string
}

const timeoutForDBPing = time.Second * 10

// SetupSuite setup at the beginning of test
func (s *Suite) SetupSuite() {
	var err error
	dbConn, err := sql.Open(postgres, s.DSN)
	s.Require().NoError(err)
	s.DBConn = dbresolver.WrapDatabaseConnection(dbConn, nil)
	pingCtx, cancel := context.WithTimeout(context.Background(), timeoutForDBPing)
	defer cancel()

	err = s.DBConn.PingContext(pingCtx)
	s.Require().NoError(err)
	_, extenErr := s.DBConn.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
	s.Require().NoError(extenErr)
	s.Migration, err = runMigration(s.DBConn.ReadWrite(), s.MigrationLocationFolder)
	s.Require().NoError(err)
}

// TearDownSuite teardown at the end of test
func (s *Suite) TearDownSuite() {
	err := s.DBConn.Close()
	s.Require().NoError(err)
}
