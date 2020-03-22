package database

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/models"
)

//Datastore is an interface that is used to inject the database into different handlers to improve testability
type Datastore interface {
	AddManualSnowdepthMeasurement(latitude, longitude, depth float64) (*models.Snowdepth, error)
	AddSnowdepthMeasurement(device *string, latitude, longitude, depth float64, when string) (*models.Snowdepth, error)
	GetLatestSnowdepths() ([]models.Snowdepth, error)
	GetLatestSnowdepthsForDevice(device string) ([]models.Snowdepth, error)
}

var dbCtxKey = &databaseContextKey{"database"}

type databaseContextKey struct {
	name string
}

// Middleware packs a pointer to the datastore into context
func Middleware(db Datastore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), dbCtxKey, db)

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

//GetFromContext extracts the database wrapper, if any, from the provided context
func GetFromContext(ctx context.Context) (Datastore, error) {
	db, ok := ctx.Value(dbCtxKey).(Datastore)
	if ok {
		return db, nil
	}

	return nil, errors.New("Failed to decode database from context")
}

type myDB struct {
	impl *gorm.DB
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

//NewDatabaseConnection initializes a new connection to the database and wraps it in a Datastore
func NewDatabaseConnection() (Datastore, error) {
	db := &myDB{}

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
			db.impl = conn
			db.impl.Debug().AutoMigrate(&models.Snowdepth{})
			break
		}
		defer conn.Close()
	}

	return db, nil
}

//AddManualSnowdepthMeasurement takes a position and a depth and adds a record to the database
func (db *myDB) AddManualSnowdepthMeasurement(latitude, longitude, depth float64) (*models.Snowdepth, error) {
	t := time.Now().UTC()
	return db.AddSnowdepthMeasurement(nil, latitude, longitude, depth, t.Format(time.RFC3339))
}

//AddSnowdepthMeasurement takes a device, position and a depth and adds a record to the database
func (db *myDB) AddSnowdepthMeasurement(device *string, latitude, longitude, depth float64, when string) (*models.Snowdepth, error) {

	measurement := &models.Snowdepth{
		Latitude:  latitude,
		Longitude: longitude,
		Depth:     float32(depth),
		Timestamp: when,
	}

	if device != nil {
		measurement.Device = *device
	}

	db.impl.Create(measurement)

	return measurement, nil
}

//GetLatestSnowdepths returns the most recent value for all sensors, as well as
//all manually added values during the last 24 hours
func (db *myDB) GetLatestSnowdepths() ([]models.Snowdepth, error) {

	// Get depths from the last 24 hours
	queryStart := time.Now().UTC().AddDate(0, 0, -1).Format(time.RFC3339)

	// TODO: Implement this as a single operation instead

	latestFromDevices := []models.Snowdepth{}
	db.impl.Table("snowdepths").Select("DISTINCT ON (device) *").Where("device <> '' AND timestamp > ?", queryStart).Order("device, timestamp desc").Find(&latestFromDevices)

	latestManual := []models.Snowdepth{}
	db.impl.Table("snowdepths").Where("device = '' AND timestamp > ?", queryStart).Find(&latestManual)

	return append(latestFromDevices, latestManual...), nil
}

func (db *myDB) GetLatestSnowdepthsForDevice(device string) ([]models.Snowdepth, error) {
	// Get depths from the last 24 hours
	queryStart := time.Now().UTC().AddDate(0, 0, -1).Format(time.RFC3339)

	depths := []models.Snowdepth{}
	db.impl.Table("snowdepths").Where("device = ? AND timestamp > ?", device, queryStart).Find(&depths)

	return depths, nil
}
