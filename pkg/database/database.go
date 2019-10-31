package database

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/models"
)

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func ConnectToDB() {

	dbHost := os.Getenv("SNOWDEPTH_DB_HOST")
	username := os.Getenv("SNOWDEPTH_DB_USER")
	dbName := os.Getenv("SNOWDEPTH_DB_NAME")
	password := os.Getenv("SNOWDEPTH_DB_PASSWORD")
	sslMode := getEnv("SNOWDEPTH_DB_SSLMODE", "require")

	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=%s password=%s", dbHost, username, dbName, sslMode, password)

	for {
		log.Printf("Connecting to database host %s ...\n", dbHost)
		conn, err := gorm.Open("postgres", dbURI)
		if err != nil {
			log.Fatalf("Failed to connect to database %s \n", err)
			time.Sleep(3 * time.Second)
		} else {
			db = conn
			db.Debug().AutoMigrate(&models.Snowdepth{})
			return
		}
		defer conn.Close()
	}
}
