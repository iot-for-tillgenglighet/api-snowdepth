package ngsi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func createURL(params ...string) string {
	url := "http://localhost:8080/ngsi-ld/v1/entities?"

	for _, p := range params {
		url = url + p + "&"
	}

	return url
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestGetEntitiesWithoutAttributesOrTypesFails(t *testing.T) {
	req, _ := http.NewRequest("GET", createURL(), nil)
	w := httptest.NewRecorder()

	NewQueryEntitiesHandler(NewContextRegistry()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Error("GET /entities MUST require either type or attrs request parameter")
	}
}

func TestGetEntitiesWithAttribute(t *testing.T) {
	req, _ := http.NewRequest("GET", createURL("attrs=snowHeight"), nil)
	w := httptest.NewRecorder()
	contextRegistry := NewContextRegistry()
	contextSource := &mockCtxSource{}

	contextRegistry.Register(contextSource)

	NewQueryEntitiesHandler(contextRegistry).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Error("That did not work .... :(")
	}
}

func TestGetEntitiesForDevice(t *testing.T) {
	req, _ := http.NewRequest("GET", createURL("attrs=snowHeight", "q=refDevice==\"urn:ngsi-ld:Device:mydevice\""), nil)
	w := httptest.NewRecorder()
	contextRegistry := NewContextRegistry()

	NewQueryEntitiesHandler(contextRegistry).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Error("That did not work ... :(")
	}
}

func newMockedContextSource(typeName string, attributeName string) *mockCtxSource {
	source := &mockCtxSource{typeName: typeName, attributeName: attributeName}
	return source
}

type mockCtxSource struct {
	typeName      string
	attributeName string
}

func (s *mockCtxSource) GetEntities(q Query, cb QueryEntitiesCallback) error {
	return nil
}

func (s *mockCtxSource) ProvidesAttribute(attributeName string) bool {
	return s.attributeName == attributeName
}

func (s *mockCtxSource) ProvidesType(typeName string) bool {
	return s.typeName == typeName
}
