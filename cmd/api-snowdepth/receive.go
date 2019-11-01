package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/database"
	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/models"
	"github.com/iot-for-tillgenglighet/messaging-golang/pkg/messaging"
)

type TelemetrySnowdepth struct {
	messaging.IoTHubMessage
	Depth float32 `json:"depth"`
}

func (t *TelemetrySnowdepth) TopicName() string {
	return "telemetry.snowdepth"
}

func receiveSnowdepth(msg amqp.Delivery) {

	log.Info("Message received from queue: " + string(msg.Body))

	depth := &TelemetrySnowdepth{}
	err := json.Unmarshal(msg.Body, depth)

	if err != nil {
		log.Error("Failed to unmarshal message")
	}

	newdepth := &models.Snowdepth{
		Device:    depth.Origin.Device,
		Latitude:  depth.Origin.Latitude,
		Longitude: depth.Origin.Longitude,
		Depth:     depth.Depth,
		Timestamp: depth.Timestamp,
	}

	database.GetDB().Create(newdepth)
}
