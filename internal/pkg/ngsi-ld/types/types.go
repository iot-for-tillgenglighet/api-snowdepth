package types

type Entity struct {
	Context []string `json:"@context"`
}

type Property struct {
	Type string `json:"type"`
}

type DateTimeProperty struct {
	Property
	Value struct {
		Type  string `json:"@type"`
		Value string `json:"@value"`
	} `json:"value"`
}

func createDateTimeProperty(value string) *DateTimeProperty {
	dtp := &DateTimeProperty{
		Property: Property{Type: "Property"},
	}

	dtp.Value.Type = "DateTime"
	dtp.Value.Value = value

	return dtp
}

type GeoJSONProperty struct {
	Property
	Value struct {
		Type        string     `json:"type"`
		Coordinates [2]float64 `json:"coordinates"`
	} `json:"value"`
}

func createGeoJSONPropertyFromWGS84(latitude, longitude float64) GeoJSONProperty {
	p := GeoJSONProperty{
		Property: Property{Type: "GeoProperty"},
	}

	p.Value.Type = "Point"
	p.Value.Coordinates[0] = longitude
	p.Value.Coordinates[1] = latitude

	return p
}

type NumberProperty struct {
	Property
	Value float64 `json:"value"`
}

func NewNumberProperty(value float64) *NumberProperty {
	return &NumberProperty{
		Property: Property{Type: "Property"},
		Value:    value,
	}
}

type TextProperty struct {
	Property
	Value string `json:"value"`
}

type Relationship struct {
	Type string `json:"type"`
}

type DeviceRelationship struct {
	Relationship
	Object string `json:"object"`
}

func createDeviceRelationshipFromDevice(device string) *DeviceRelationship {
	if len(device) == 0 {
		return nil
	}

	return &DeviceRelationship{
		Relationship: Relationship{Type: "Relationship"},
		Object:       "urn:ngsi-ld:Device:" + device,
	}
}

type WeatherObserved struct {
	ID           string              `json:"id"`
	Type         string              `json:"type"`
	DateCreated  *DateTimeProperty   `json:"dateCreated,omitempty"`
	DateModified *DateTimeProperty   `json:"dateModified,omitempty"`
	DateObserved DateTimeProperty    `json:"dateObserved"`
	Location     GeoJSONProperty     `json:"location"`
	RefDevice    *DeviceRelationship `json:"refDevice,omitempty"`
	SnowHeight   *NumberProperty     `json:"snowHeight,omitempty"`
	Entity
}

func NewWeatherObserved(device string, latitude float64, longitude float64, observedAt string) *WeatherObserved {
	dateTimeValue := createDateTimeProperty(observedAt)
	refDevice := createDeviceRelationshipFromDevice(device)

	if refDevice == nil {
		device = "manual"
	}

	id := "urn:ngsi-ld:WeatherObserved:SnowHeight:" + device + ":" + observedAt

	return &WeatherObserved{
		ID:           id,
		Type:         "WeatherObserved",
		DateObserved: *dateTimeValue,
		Location:     createGeoJSONPropertyFromWGS84(latitude, longitude),
		RefDevice:    refDevice,
		Entity: Entity{
			Context: []string{
				"https://schema.lab.fiware.org/ld/context",
				"https://uri.etsi.org/ngsi-ld/v1/ngsi-ld-core-context.jsonld",
			},
		},
	}
}
