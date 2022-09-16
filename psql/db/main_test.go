package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/bwen19/blog/util"
)

// var testQueries *Queries
// var testDB *sql.DB
var testStore Store

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	testDB, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// testQueries = New(testDB)
	testStore = NewStore(testDB)

	os.Exit(m.Run())
}
