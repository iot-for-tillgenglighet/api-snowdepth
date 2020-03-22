package main

import (
	log "github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/database"
	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/handler"
	"github.com/iot-for-tillgenglighet/messaging-golang/pkg/messaging"
	"github.com/iot-for-tillgenglighet/messaging-golang/pkg/messaging/telemetry"
)

func main() {

	serviceName := "api-snowdepth"

	log.Infof("Starting up %s ...", serviceName)

	config := messaging.LoadConfiguration(serviceName)
	messenger, _ := messaging.Initialize(config)

	defer messenger.Close()

	db, _ := database.NewDatabaseConnection()

	messenger.RegisterTopicMessageHandler((&telemetry.Snowdepth{}).TopicName(), createSnowdepthReceiver(db))

	handler.CreateRouterAndStartServing(db)
}
