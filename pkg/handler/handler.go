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
	"github.com/rs/cors"

	log "github.com/sirupsen/logrus"
)

func Router() {

	router := chi.NewRouter()

	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	// Enable gzip compression for ngsi-ld responses
	compressor := middleware.NewCompressor(flate.DefaultCompression, "application/json", "application/ld+json")
	router.Use(compressor.Handler())

	router.Use(middleware.Logger)

	gqlServer := handler.New(gql.NewExecutableSchema(gql.Config{Resolvers: &gql.Resolver{}}))
	gqlServer.AddTransport(&transport.POST{})
	gqlServer.Use(extension.Introspection{})

	router.Handle("/api/graphql/playground", playground.Handler("GraphQL playground", "/api/graphql"))
	router.Handle("/api/graphql", gqlServer)

	router.Get("/ngsi-ld/v1/entities", ngsi.QueryEntities)

	port := os.Getenv("SNOWDEPTH_API_PORT")
	if port == "" {
		port = "8880"
	}

	log.Printf("Starting api-snowdepth on port %s.\n", port)

	log.Fatal(http.ListenAndServe(":"+port, router))
}
