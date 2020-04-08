package ngsi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterContextSource(t *testing.T) {
	registrationBody := NewCsourceRegistration("Point", []string{"x", "y"}, "lolcathost")
	jsonBytes, _ := json.Marshal(registrationBody)
	ctxRegistry := NewContextRegistry()
	req, _ := http.NewRequest("POST", createURL("/csourceRegistration"), bytes.NewBuffer(jsonBytes))
	w := httptest.NewRecorder()

	NewRegisterContextSourceHandler(ctxRegistry).ServeHTTP(w, req)

	sources := ctxRegistry.GetContextSourcesForQuery(
		newQueryFromParameters(nil, []string{"Point"}, []string{"x"}, ""),
	)

	if len(sources) != 1 {
		t.Error("The registered context source was not added to the registry.")
	}

	if w.Code != http.StatusCreated {
		t.Error("Wrong status code returned. ", w.Code, " != expected 201")
	}
}

func TestThatRequestsAreForwardedToRemoteContext(t *testing.T) {
	mockService := setupMockServiceThatReturns(200, snowHeightResponseJSON)
	defer mockService.Close()

	remoteURL := mockService.URL
	registrationBody := NewCsourceRegistration("WeatherObserved", []string{"snowHeight"}, remoteURL)
	jsonBytes, _ := json.Marshal(registrationBody)
	ctxRegistry := NewContextRegistry()

	// Send a POST request to register a remote context source
	req, _ := http.NewRequest("POST", createURL("/csourceRegistration"), bytes.NewBuffer(jsonBytes))
	w := httptest.NewRecorder()
	NewRegisterContextSourceHandler(ctxRegistry).ServeHTTP(w, req)

	// Send a GET request for entities of type WeatherObserved (that are handled by the "remote" source)
	req, _ = http.NewRequest("GET", "https://localhost/ngsi-ld/v1/entities?type=WeatherObserved", nil)
	query := newQueryFromParameters(req, []string{"WeatherObserved"}, []string{"snowHeight"}, "")
	sources := ctxRegistry.GetContextSourcesForQuery(query)

	numEntities := 0

	for _, src := range sources {
		src.GetEntities(query, func(entity Entity) error {
			numEntities++
			return nil
		})
	}

	if numEntities == 0 {
		t.Error("Failed to get entities from remote endpoint.")
	}
}

const snowHeightResponseJSON string = "[{\"id\": \"urn:ngsi-ld:WeatherObserved:SnowHeight:snow_10a52aaa84c35727:2020-04-08T15:01:32Z\", \"type\": \"WeatherObserved\",\"dateObserved\": { \"type\": \"Property\", \"value\": {\"@type\": \"DateTime\", \"@value\": \"2020-04-08T15:01:32Z\"}}, \"location\": { \"type\": \"GeoProperty\", \"value\": { \"type\": \"Point\", \"coordinates\": [16.5687632, 62.4081681]}}, \"refDevice\": {\"type\": \"Relationship\", \"object\": \"urn:ngsi-ld:Device:snow_10a52aaa84c35727\"}, \"snowHeight\": { \"type\": \"Property\", \"value\": 0}, \"@context\": [\"https://schema.lab.fiware.org/ld/context\", \"https://uri.etsi.org/ngsi-ld/v1/ngsi-ld-core-context.jsonld\"]}]"

func setupMockServiceThatReturns(responseCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(responseCode)
		w.Header().Add("Content-Type", "application/ld+json")
		if body != "" {
			w.Write([]byte(body))
		}
	}))
}
