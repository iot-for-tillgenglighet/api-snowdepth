// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package graphql

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql/introspection"
)

func (ec *executionContext) __resolve__service(ctx context.Context) (introspection.Service, error) {
	if ec.DisableIntrospection {
		return introspection.Service{}, errors.New("federated introspection disabled")
	}
	return introspection.Service{
		SDL: `scalar DateTime
type Device @key(fields: "id") {
	id: ID!
}
input MeasurementPosition {
	lon: Float!
	lat: Float!
}
type Mutation @extends {
	addSnowdepthMeasurement(input: NewSnowdepthMeasurement!): Snowdepth!
}
input NewSnowdepthMeasurement {
	pos: MeasurementPosition!
	depth: Float!
}
type Origin {
	device: Device
	pos: WGS84Position
}
type Query @extends {
	snowdepths: [Snowdepth]!
}
type Snowdepth implements Telemetry {
	from: Origin!
	when: DateTime!
	depth: Float!
	manual: Boolean
}
interface Telemetry {
	from: Origin!
	when: DateTime!
}
type WGS84Position {
	lon: Float!
	lat: Float!
}
`,
	}, nil
}

func (ec *executionContext) __resolve_entities(ctx context.Context, representations []map[string]interface{}) ([]_Entity, error) {
	list := []_Entity{}
	for _, rep := range representations {
		typeName, ok := rep["__typename"].(string)
		if !ok {
			return nil, errors.New("__typename must be an existing string")
		}
		switch typeName {

		case "Device":
			id, ok := rep["id"].(string)
			if !ok {
				return nil, errors.New("opsies")
			}
			resp, err := ec.resolvers.Entity().FindDeviceByID(ctx, id)
			if err != nil {
				return nil, err
			}

			list = append(list, resp)

		default:
			return nil, errors.New("unknown type: " + typeName)
		}
	}
	return list, nil
}
