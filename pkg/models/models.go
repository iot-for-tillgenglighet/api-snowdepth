package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Snowdepth struct {
	gorm.Model
	Latitude  float64
	Longitude float64
	Device    string `gorm:"unique_index:idx_device_timestamp"`
	Depth     float32
	Timestamp string `gorm:"unique_index:idx_device_timestamp"`
}
