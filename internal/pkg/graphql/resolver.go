// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.
package graphql

import (
	"context"
	"math"

	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/database"
	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/models"
)

type Resolver struct{}

func (r *entityResolver) FindDeviceByID(ctx context.Context, id string) (*Device, error) {
	return &Device{ID: id}, nil
}

func convertDatabaseRecordToGQL(measurement *models.Snowdepth) *Snowdepth {
	if measurement != nil {
		depth := &Snowdepth{
			From: &Origin{
				Pos: &WGS84Position{
					Lat: measurement.Latitude,
					Lon: measurement.Longitude,
				},
			},
			When:  measurement.Timestamp,
			Depth: math.Round(float64(measurement.Depth*10)) / 10,
		}

		if len(measurement.Device) == 0 {
			depth.Manual = &[]bool{true}[0] // <- You may Google that little nugget of beauty ...
		} else {
			depth.Manual = &[]bool{false}[0]
			depth.From.Device = &Device{ID: measurement.Device}
		}

		return depth
	}

	return nil
}

func (r *mutationResolver) AddSnowdepthMeasurement(ctx context.Context, input NewSnowdepthMeasurement) (*Snowdepth, error) {
	measurement, err := database.AddManualSnowdepthMeasurement(input.Pos.Lat, input.Pos.Lon, input.Depth)
	return convertDatabaseRecordToGQL(measurement), err
}

func (r *queryResolver) Snowdepths(ctx context.Context) ([]*Snowdepth, error) {
	depths, err := database.GetLatestSnowdepths()

	if err != nil {
		panic("Failed to query latest snowdepths: " + err.Error())
	}

	depthcount := len(depths)

	if depthcount == 0 {
		return []*Snowdepth{}, nil
	}

	gqldepths := make([]*Snowdepth, 0, depthcount)

	for _, v := range depths {
		gqldepths = append(gqldepths, convertDatabaseRecordToGQL(&v))
	}

	return gqldepths, nil
}

func (r *Resolver) Entity() EntityResolver     { return &entityResolver{r} }
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }
func (r *Resolver) Query() QueryResolver       { return &queryResolver{r} }

type entityResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
