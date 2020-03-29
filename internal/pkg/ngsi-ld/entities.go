package ngsi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iot-for-tillgenglighet/api-snowdepth/internal/pkg/ngsi-ld/errors"
	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/models"
)

//Entity is an informational representative of something that is supposed to exist in the real world, physically or conceptually
type Entity interface {
}

//entitiesDatastore is an interface containing all the entities related datastore functions
type entitiesDatastore interface {
	GetLatestSnowdepths() ([]models.Snowdepth, error)
	GetLatestSnowdepthsForDevice(device string) ([]models.Snowdepth, error)
}

type QueryEntitiesCallback func(entity Entity) error

//NewQueryEntitiesHandler handles GET requests for NGSI entitites
func NewQueryEntitiesHandler(ctxReg ContextRegistry) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entityTypeNames := r.URL.Query().Get("type")
		attributeNames := r.URL.Query().Get("attrs")

		if entityTypeNames == "" && attributeNames == "" {
			errors.ReportNewInvalidRequest(
				w,
				"A request for entities MUST specify at least one of type and attrs.",
			)
			return
		}

		entityTypes := strings.Split(entityTypeNames, ",")
		attributes := strings.Split(attributeNames, ",")

		q := r.URL.Query().Get("q")
		query := newQueryFromParameters(entityTypes, attributes, q)

		contextSources := ctxReg.GetContextSourcesForQuery(query)

		var entities = []Entity{}
		var err error

		// TODO: Iterate over the context sources and concatenate the results
		if len(contextSources) > 0 {
			err = contextSources[0].GetEntities(query, func(entity Entity) error {
				entities = append(entities, entity)
				return nil
			})
		}

		if err != nil {
			errors.ReportNewInternalError(
				w,
				"An internal error was encountered when trying to get entities from the database.",
			)
			return
		}

		bytes, err := json.MarshalIndent(entities, "", "  ")
		if err != nil {
			errors.ReportNewInternalError(w, "Failed to encode response.")
			return
		}

		w.Header().Add("Content-Type", "application/ld+json")
		w.Write(bytes)
	})
}
