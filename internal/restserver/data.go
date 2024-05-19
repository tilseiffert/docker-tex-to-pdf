package restserver

import "gorm.io/gorm"

type Jobs struct {
	gorm.Model
	JobID         string `json:"ulid"` // ULID, index
	Name          string `json:"name"`
	Status        string `json:"status"`
	StatusRunning bool   `json:"status_running"`
	StatusSuccess bool   `json:"status_success"`
	Error         string `json:"error"`  // any error message
	Path          string `json:"path"`   // absolute path to the build dir
	Result        string `json:"result"` // absolute path to the resulting PDF/A file
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Jobs{})
}
