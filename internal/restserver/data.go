package restserver

import "gorm.io/gorm"

type Jobs struct {
	gorm.Model
	JobID       string `json:"job_id"` // ULID
	JobStatus   string `json:"job_status"`
	JobProgress int    `json:"job_progress"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Jobs{})
}
