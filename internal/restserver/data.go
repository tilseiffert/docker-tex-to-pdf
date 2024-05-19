package restserver

import "gorm.io/gorm"

type Jobs struct {
	gorm.Model
	JobID  string `json:"ulid"` // ULID, index
	Name   string `json:"name"`
	Status string `json:"status"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Jobs{})
}
