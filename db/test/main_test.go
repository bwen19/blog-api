package test

import (
	"blog/server/db/sqlc"
	"blog/server/util"
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v4"
)

var testQueries *sqlc.Queries
var testDB *pgx.Conn

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	testDB, err = pgx.Connect(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = sqlc.New(testDB)

	os.Exit(m.Run())
}
