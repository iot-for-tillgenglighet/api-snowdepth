package handler

import (
	"compress/flate"
	"errors"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	gql "github.com/iot-for-tillgenglighet/api-snowdepth/internal/pkg/graphql"
	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/database"
	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/models"
	"github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/datamodels/fiware"
	ngsi "github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/ngsi-ld"
	"github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/ngsi-ld/types"
	"github.com/rs/cors"

	log "github.com/sirupsen/logrus"
)

//RequestRouter wraps the concrete router implementation
type RequestRouter struct {
	impl *chi.Mux
}

func (router *RequestRouter) addGraphQLHandlers(db database.Datastore) {
	gqlServer := handler.New(gql.NewExecutableSchema(gql.Config{Resolvers: &gql.Resolver{}}))
	gqlServer.AddTransport(&transport.POST{})
	gqlServer.Use(extension.Introspection{})

	// TODO: Investigate some way to use closures instead of context even for GraphQL handlers
	router.impl.Use(database.Middleware(db))

	router.impl.Handle("/api/graphql/playground", playground.Handler("GraphQL playground", "/api/graphql"))
	router.impl.Handle("/api/graphql", gqlServer)
}

func (router *RequestRouter) addNGSIHandlers(contextRegistry ngsi.ContextRegistry) {
	router.Get("/ngsi-ld/v1/entities", ngsi.NewQueryEntitiesHandler(contextRegistry))
	router.Get("/ngsi-ld/v1/entities/{entity}", ngsi.NewRetrieveEntityHandler(contextRegistry))
	router.Patch("/ngsi-ld/v1/entities/{entity}/attrs/", ngsi.NewUpdateEntityAttributesHandler(contextRegistry))
	router.Post("/ngsi-ld/v1/entities", ngsi.NewCreateEntityHandler(contextRegistry))
}

func (router *RequestRouter) addProbeHandlers() {
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

//Get accepts a pattern that should be routed to the handlerFn on a GET request
func (router *RequestRouter) Get(pattern string, handlerFn http.HandlerFunc) {
	router.impl.Get(pattern, handlerFn)
}

//Patch accepts a pattern that should be routed to the handlerFn on a PATCH request
func (router *RequestRouter) Patch(pattern string, handlerFn http.HandlerFunc) {
	router.impl.Patch(pattern, handlerFn)
}

//Post accepts a pattern that should be routed to the handlerFn on a POST request
func (router *RequestRouter) Post(pattern string, handlerFn http.HandlerFunc) {
	router.impl.Post(pattern, handlerFn)
}

//newRequestRouter creates and returns a new router wrapper
func newRequestRouter() *RequestRouter {
	router := &RequestRouter{impl: chi.NewRouter()}

	router.impl.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	// Enable gzip compression for ngsi-ld responses
	compressor := middleware.NewCompressor(flate.DefaultCompression, "application/json", "application/ld+json")
	router.impl.Use(compressor.Handler)
	router.impl.Use(middleware.Logger)

	return router
}

func createRequestRouter(contextRegistry ngsi.ContextRegistry, db database.Datastore) *RequestRouter {
	router := newRequestRouter()

	router.addGraphQLHandlers(db)
	router.addNGSIHandlers(contextRegistry)
	router.addProbeHandlers()

	return router
}

//CreateRouterAndStartServing creates a request router, registers all handlers and starts serving requests
func CreateRouterAndStartServing(db database.Datastore) {

	contextRegistry := ngsi.NewContextRegistry()
	ctxSource := contextSource{db: db}
	contextRegistry.Register(ctxSource)

	// Enable this code to allow the snowdepth service to do double duty as a broker in the iot-hub
	remoteURL := "http://api-problemreport-service.iot.svc.cluster.local/"
	registration, _ := ngsi.NewCsourceRegistration("Open311ServiceRequest", []string{"service_code"}, remoteURL, nil)
	contextSource, _ := ngsi.NewRemoteContextSource(registration)
	contextRegistry.Register(contextSource)

	remoteURL = "http://api-temperature-service.iot.svc.cluster.local/"
	registration, _ = ngsi.NewCsourceRegistration("WeatherObserved", []string{"temperature"}, remoteURL, nil)
	contextSource, _ = ngsi.NewRemoteContextSource(registration)
	contextRegistry.Register(contextSource)

	remoteURL = "http://api-temperature-service.iot.svc.cluster.local/"
	registration, _ = ngsi.NewCsourceRegistration("WaterQualityObserved", []string{"temperature"}, remoteURL, nil)
	contextSource, _ = ngsi.NewRemoteContextSource(registration)
	contextRegistry.Register(contextSource)

	remoteURL = "http://api-transportation-service.iot.svc.cluster.local/"
	regex := "^urn:ngsi-ld:Road:.+"
	registration, _ = ngsi.NewCsourceRegistration("Road", []string{}, remoteURL, &regex)
	contextSource, _ = ngsi.NewRemoteContextSource(registration)
	contextRegistry.Register(contextSource)

	remoteURL = "http://api-transportation-service.iot.svc.cluster.local/"
	regex = "^urn:ngsi-ld:RoadSegment:.+"
	registration, _ = ngsi.NewCsourceRegistration("RoadSegment", []string{}, remoteURL, &regex)
	contextSource, _ = ngsi.NewRemoteContextSource(registration)
	contextRegistry.Register(contextSource)

	remoteURL = "http://iot-device-registry-service.iot.svc.cluster.local/"
	regex = "^urn:ngsi-ld:Device:.+"
	registration, _ = ngsi.NewCsourceRegistration("Device", []string{"value"}, remoteURL, &regex)
	contextSource, _ = ngsi.NewRemoteContextSource(registration)
	contextRegistry.Register(contextSource)

	remoteURL = "http://iot-device-registry-service.iot.svc.cluster.local/"
	regex = "^urn:ngsi-ld:DeviceModel:.+"
	registration, _ = ngsi.NewCsourceRegistration("DeviceModel", []string{}, remoteURL, &regex)
	contextSource, _ = ngsi.NewRemoteContextSource(registration)
	contextRegistry.Register(contextSource)

	router := createRequestRouter(contextRegistry, db)

	port := os.Getenv("SNOWDEPTH_API_PORT")
	if port == "" {
		port = "8880"
	}

	log.Printf("Starting api-snowdepth on port %s.\n", port)

	log.Fatal(http.ListenAndServe(":"+port, router.impl))
}

type contextSource struct {
	db database.Datastore
}

func convertDatabaseRecordToWeatherObserved(r *models.Snowdepth) *fiware.WeatherObserved {
	if r != nil {
		entity := fiware.NewWeatherObserved("snowHeight:"+r.Device, r.Latitude, r.Longitude, r.Timestamp)
		entity.SnowHeight = types.NewNumberProperty(math.Round(float64(r.Depth*10)) / 10)
		return entity
	}

	return nil
}

func (cs contextSource) CreateEntity(typeName, entityID string, req ngsi.Request) error {
	return nil
}

func (cs contextSource) GetEntities(query ngsi.Query, callback ngsi.QueryEntitiesCallback) error {

	var snowdepths []models.Snowdepth
	var err error

	if query.HasDeviceReference() {
		deviceID := strings.TrimPrefix(query.Device(), "urn:ngsi-ld:Device:")
		snowdepths, err = cs.db.GetLatestSnowdepthsForDevice(deviceID)
	} else {
		snowdepths, err = cs.db.GetLatestSnowdepths()
	}

	if err == nil {
		for _, v := range snowdepths {
			err = callback(convertDatabaseRecordToWeatherObserved(&v))
			if err != nil {
				break
			}
		}
	}

	return err
}

func (cs contextSource) ProvidesAttribute(attributeName string) bool {
	return attributeName == "snowHeight"
}

func (cs contextSource) ProvidesEntitiesWithMatchingID(entityID string) bool {
	// not supported yet
	return false
}

func (cs contextSource) ProvidesType(typeName string) bool {
	return typeName == "WeatherObserved"
}

func (cs contextSource) UpdateEntityAttributes(entityID string, req ngsi.Request) error {
	return errors.New("UpdateEntityAttributes is not supported")
}

func (cs contextSource) RetrieveEntity(entityID string, req ngsi.Request) (ngsi.Entity, error) {
	return nil, errors.New("RetrieveEntity is not supported")
}
