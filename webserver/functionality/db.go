package functionality

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)
var DB *gorm.DB
// Initialize database connection using GORM + SQLite
func InitDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("webserver.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&User{}, &Profile{})
	if err != nil {
		log.Fatal("failed to migrate database:", err)
	}

	return db
}