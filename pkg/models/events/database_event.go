package events

import (
	"fmt"
	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/jackc/pgtype"
	"gorm.io/gorm"
	"time"
)

type DatabaseEvent struct {
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	EventID       string
	CorrelationID string       `gorm:"index"`
	Timestamp     time.Time    `gorm:"index"`
	SourceName    string       `gorm:"index"`
	SourceID      string       `gorm:"index"`
	TypeName      string       `gorm:"index"`
	Data          pgtype.JSONB `gorm:"type:jsonb"`
	IsHandled     bool         `gorm:"index"`
	HandledByName string       `gorm:"index"`
	HandledByID   string       `gorm:"index"`
}

func (d *DatabaseEvent) TableName() string {
	return tableName
}

func (d *DatabaseEvent) Create(givenDB *gorm.DB) (*gorm.DB, error) {
	returnedDB := givenDB.Create(d)

	return returnedDB, returnedDB.Error
}

func (d *DatabaseEvent) Update(givenDB *gorm.DB) (*gorm.DB, error) {
	returnedDB := givenDB.Model(DatabaseEvent{}).Where("event_id = ? AND created_at = ?", d.EventID, d.CreatedAt).Updates(d)

	return returnedDB, returnedDB.Error
}

func (d *DatabaseEvent) Delete(givenDB *gorm.DB) (*gorm.DB, error) {
	returnedDB := givenDB.Model(DatabaseEvent{}).Where("event_id = ? AND created_at = ?", d.EventID, d.CreatedAt).Delete(d)

	return returnedDB, returnedDB.Error
}

func Migrate(db *gorm.DB) error {
	var err error

	err = db.AutoMigrate(&DatabaseEvent{})
	if err != nil {
		return err
	}

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

func GetAll(db *gorm.DB) ([]*DatabaseEvent, error) {
	rows := make([]*DatabaseEvent, 0)

	returnedDB := db.Find(&rows)

	return rows, returnedDB.Error
}
