package ngsi

import (
	"encoding/json"
	"math"
	"net/http"

	"github.com/iot-for-tillgenglighet/api-snowdepth/internal/pkg/ngsi-ld/errors"
	"github.com/iot-for-tillgenglighet/api-snowdepth/internal/pkg/ngsi-ld/types"
	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/database"
	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/models"
)

func convertDatabaseRecordToWeatherObserved(r *models.Snowdepth) *types.WeatherObserved {
	if r != nil {
		entity := types.NewWeatherObserved(r.Device, r.Latitude, r.Longitude, r.Timestamp)
		entity.SnowHeight = types.NewNumberProperty(math.Round(float64(r.Depth*10)) / 10)
		return entity
	}

	return nil
}

//QueryEntities handles GET requests for NGSI entitites
func QueryEntities(w http.ResponseWriter, r *http.Request) {
	entityTypes := r.URL.Query().Get("type")

	if entityTypes != "" && entityTypes != "WeatherObserved" {
		errors.ReportNewInvalidRequest(w, "Entity type not supported by this service")
		return
	}

	snowdepths, err := database.GetLatestSnowdepths()

	if err != nil {
		errors.ReportNewInternalError(w, "An internal error was encountered when trying to get entities from the database.")
		return
	}

	depthcount := len(snowdepths)
	ngsiEntities := []*types.WeatherObserved{}

	if depthcount > 0 {
		ngsiEntities = make([]*types.WeatherObserved, 0, depthcount)

		for _, v := range snowdepths {
			ngsiEntities = append(ngsiEntities, convertDatabaseRecordToWeatherObserved(&v))
		}
	}

	bytes, err := json.MarshalIndent(ngsiEntities, "", "  ")
	if err != nil {
		errors.ReportNewInternalError(w, "Failed to encode response.")
		return
	}

	w.Header().Add("Content-Type", "application/ld+json")
	w.Write(bytes)
}
