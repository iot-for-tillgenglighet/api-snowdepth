package main

import (
	"time"

	log "github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/database"
	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/handler"
)

func main() {

	log.Info("Starting api-snowdepth")

	time.Sleep(30 * time.Second)

	database.ConnectToDB()

	connection, channel := receiveSnowdepth()

	defer connection.Close()
	defer channel.Close()

	handler.Router()
}
