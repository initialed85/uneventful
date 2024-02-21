package states

import (
	"fmt"
	"time"

	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/jackc/pgtype"
	"gorm.io/gorm"
)

type DatabaseState struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	VersionID uint64         `gorm:"autoIncrement"` // unique
	Timestamp time.Time      `gorm:"index"`
	Name      string         `gorm:"index"`
	EntityID  string         `gorm:"index"`
	Data      pgtype.JSONB   `gorm:"type:jsonb"`
}

func (d *DatabaseState) TableName() string {
	return tableName
}

func (d *DatabaseState) Create(givenDB *gorm.DB) (*gorm.DB, error) {
	returnedDB := givenDB.Create(d)

	return returnedDB, returnedDB.Error
}

func (d *DatabaseState) Update(givenDB *gorm.DB) (*gorm.DB, error) {
	returnedDB := givenDB.Model(DatabaseState{}).Where("version_id = ? AND created_at = ?", d.VersionID, d.CreatedAt).Updates(d)

	return returnedDB, returnedDB.Error
}

func (d *DatabaseState) Delete(givenDB *gorm.DB) (*gorm.DB, error) {
	returnedDB := givenDB.Model(DatabaseState{}).Where("version_id = ? AND created_at = ?", d.VersionID, d.CreatedAt).Delete(d)

	return returnedDB, returnedDB.Error
}

func Migrate(db *gorm.DB) error {
	var err error

	err = db.AutoMigrate(&DatabaseState{})
	if err != nil {
		return err
	}

	db.Exec(fmt.Sprintf(createUniqueIndexSQL, tableName))

	db.Exec(fmt.Sprintf(createUniqueIndexSQL, tableName))

	useSQLite, err := helpers.GetEnvironmentVariable("USE_SQLITE", false, "0")
	if err != nil {
		return err
	}

	if useSQLite != "1" {
		db.Exec(fmt.Sprintf(createdHypertableSQL, tableName))
	}

	return nil
}

func GetAll(db *gorm.DB) ([]*DatabaseState, error) {
	rows := make([]*DatabaseState, 0)

	returnedDB := db.Find(&rows)

	return rows, returnedDB.Error
}
