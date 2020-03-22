package ngsi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/models"
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

func TestGetEntities(t *testing.T) {
	req, _ := http.NewRequest("GET", createURL("attrs=snowHeight"), nil)
	w := httptest.NewRecorder()
	db := &mockDatastore{}

	NewQueryEntitiesHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Error("That did not work ... :(")
	}
}

func TestGetEntitiesForDevice(t *testing.T) {
	req, _ := http.NewRequest("GET", createURL("attrs=snowHeight", "q=refDevice==\"urn:ngsi-ld:Device:mydevice\""), nil)
	w := httptest.NewRecorder()
	db := &mockDatastore{}

	NewQueryEntitiesHandler(db).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Error("That did not work ... :(")
	}
}

type mockDatastore struct {
	entities map[string][]*models.Snowdepth
}

func (db *mockDatastore) GetLatestSnowdepths() ([]models.Snowdepth, error) {
	return []models.Snowdepth{}, nil
}

func (db *mockDatastore) GetLatestSnowdepthsForDevice(device string) ([]models.Snowdepth, error) {
	if device == "mydevice" {
		e := responseEntity(device)
		return []models.Snowdepth{e}, nil
	}

	return []models.Snowdepth{}, nil
}

func responseEntity(device string) models.Snowdepth {
	depth := models.Snowdepth{
		Device: device,
	}
	return depth
}
