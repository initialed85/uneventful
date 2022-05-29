package helpers

import (
	"fmt"
	"github.com/initialed85/uneventful/internal/constants"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func GetDatabase() (db *gorm.DB, err error) {
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

	dsn := fmt.Sprintf(
		"host=%v port=%v user=%v password=%v dbname=%v sslmode=disable TimeZone=Australia/Perth",
		postgresHost, postgresPort, postgresUser, postgresPassword, postgresDatabase,
	)

	db, err = gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
