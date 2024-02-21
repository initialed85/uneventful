package helpers

import (
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"github.com/initialed85/uneventful/internal/constants"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetDatabase() (db *gorm.DB, err error) {
	useSQLite, err := GetEnvironmentVariable("USE_SQLITE", false, "0")
	if err != nil {
		log.Fatal(err)
	}

	if useSQLite == "1" {
		db, err = gorm.Open(sqlite.Open("/var/lib/sqlite/data/datastore.db"), &gorm.Config{})
	} else {
		postgresHost, err := GetEnvironmentVariable("POSTGRES_HOST", true, "")
		if err != nil {
			log.Fatal(err)
		}

		postgresPort, err := GetEnvironmentVariable("POSTGRES_PORT", false, constants.DefaultPostgresPort)
		if err != nil {
			log.Fatal(err)
		}

		postgresUser, err := GetEnvironmentVariable("POSTGRES_USER", false, constants.DefaultPostgresUser)
		if err != nil {
			log.Fatal(err)
		}

		postgresPassword, err := GetEnvironmentVariable("POSTGRES_PASSWORD", false, constants.DefaultPostgresPassword)
		if err != nil {
			log.Fatal(err)
		}

		postgresDatabase, err := GetEnvironmentVariable("POSTGRES_DB", false, constants.DefaultPostgresDatabase)
		if err != nil {
			log.Fatal(err)
		}

		dsn := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable TimeZone=Australia/Perth", postgresHost, postgresPort, postgresUser, postgresPassword, postgresDatabase)

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}
