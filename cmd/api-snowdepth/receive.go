package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/database"
	"github.com/iot-for-tillgenglighet/messaging-golang/pkg/messaging/telemetry"
)

func receiveSnowdepth(msg amqp.Delivery) {

	log.Info("Message received from queue: " + string(msg.Body))

	depth := &telemetry.Snowdepth{}
	err := json.Unmarshal(msg.Body, depth)

	if err != nil {
		log.Error("Failed to unmarshal message")
		return
	}

	// TODO: Propagate database errors, catch and log them here ...
	database.AddSnowdepthMeasurement(
		&depth.Origin.Device,
		depth.Origin.Latitude, depth.Origin.Longitude,
		float64(depth.Depth),
		depth.Timestamp,
	)
}
