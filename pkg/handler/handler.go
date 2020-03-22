package handler

import (
	"compress/flate"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	gql "github.com/iot-for-tillgenglighet/api-snowdepth/internal/pkg/graphql"
	ngsi "github.com/iot-for-tillgenglighet/api-snowdepth/internal/pkg/ngsi-ld"
	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/database"
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

func (router *RequestRouter) addNGSIHandlers(db database.Datastore) {
	router.Get("/ngsi-ld/v1/entities", ngsi.NewQueryEntitiesHandler(db))
}

//Get accepts a pattern that should be routed to the handlerFn on a GET request
func (router *RequestRouter) Get(pattern string, handlerFn http.HandlerFunc) {
	router.impl.Get(pattern, handlerFn)
}

//NewRequestRouter creates and returns a new router wrapper
func NewRequestRouter() *RequestRouter {
	router := &RequestRouter{impl: chi.NewRouter()}

	router.impl.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	// Enable gzip compression for ngsi-ld responses
	compressor := middleware.NewCompressor(flate.DefaultCompression, "application/json", "application/ld+json")
	router.impl.Use(compressor.Handler())
	router.impl.Use(middleware.Logger)

	return router
}

func createRequestRouter(db database.Datastore) *RequestRouter {
	router := NewRequestRouter()

	router.addGraphQLHandlers(db)
	router.addNGSIHandlers(db)

	return router
}

//CreateRouterAndStartServing creates a request router, registers all handlers and starts serving requests
func CreateRouterAndStartServing(db database.Datastore) {

	router := createRequestRouter(db)

	port := os.Getenv("SNOWDEPTH_API_PORT")
	if port == "" {
		port = "8880"
	}

	log.Printf("Starting api-snowdepth on port %s.\n", port)

	log.Fatal(http.ListenAndServe(":"+port, router.impl))
}
