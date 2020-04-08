package ngsi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func createURL(path string, params ...string) string {
	url := "http://localhost:8080/ngsi-ld/v1" + path

	if len(params) > 0 {
		url = url + "?"

		for _, p := range params {
			url = url + p + "&"
		}

		url = strings.TrimSuffix(url, "&")
	}

	return url
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestGetEntitiesWithoutAttributesOrTypesFails(t *testing.T) {
	req, _ := http.NewRequest("GET", createURL("/entitites"), nil)
	w := httptest.NewRecorder()

	NewQueryEntitiesHandler(NewContextRegistry()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Error("GET /entities MUST require either type or attrs request parameter")
	}
}

func TestGetEntitiesWithAttribute(t *testing.T) {
	req, _ := http.NewRequest("GET", createURL("/entitites", "attrs=snowHeight"), nil)
	w := httptest.NewRecorder()
	contextRegistry := NewContextRegistry()
	contextRegistry.Register(newMockedContextSource(
		"", "snowHeight",
		e(""),
	))

	NewQueryEntitiesHandler(contextRegistry).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Error("That did not work .... :(")
	}
}

func TestGetEntitiesForDevice(t *testing.T) {
	deviceID := "urn:ngsi-ld:Device:mydevice"
	req, _ := http.NewRequest("GET", createURL("/entitites", "attrs=snowHeight", "q=refDevice==\""+deviceID+"\""), nil)
	w := httptest.NewRecorder()
	contextRegistry := NewContextRegistry()
	contextSource := newMockedContextSource("", "snowHeight")
	contextRegistry.Register(contextSource)

	NewQueryEntitiesHandler(contextRegistry).ServeHTTP(w, req)

	if contextSource.queriedDevice != deviceID {
		t.Error("Queried device did not match expectations. ", contextSource.queriedDevice, " != ", deviceID)
	} else if w.Code != http.StatusOK {
		t.Error("That did not work ... :(")
	}
}

type mockEntity struct {
	Value string
}

func e(val string) mockEntity {
	return mockEntity{Value: val}
}

func newMockedContextSource(typeName string, attributeName string, e ...mockEntity) *mockCtxSource {
	source := &mockCtxSource{typeName: typeName, attributeName: attributeName}
	for _, entity := range e {
		source.entities = append(source.entities, entity)
	}
	return source
}

type mockCtxSource struct {
	typeName      string
	attributeName string
	entities      []Entity

	queriedDevice string
}

func (s *mockCtxSource) GetEntities(q Query, cb QueryEntitiesCallback) error {

	if q.HasDeviceReference() {
		s.queriedDevice = q.Device()
	}

	for _, e := range s.entities {
		cb(e)
	}
	return nil
}

func (s *mockCtxSource) ProvidesAttribute(attributeName string) bool {
	return s.attributeName == attributeName
}

func (s *mockCtxSource) ProvidesType(typeName string) bool {
	return s.typeName == typeName
}
