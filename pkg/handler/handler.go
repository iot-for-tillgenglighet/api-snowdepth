package handler

import (
	"net/http"
	"os"

	"github.com/99designs/gqlgen/handler"
	gql "github.com/iot-for-tillgenglighet/api-snowdepth/internal/pkg/graphql"

	log "github.com/sirupsen/logrus"
)

func Router() {

	port := os.Getenv("SNOWDEPTH_API_PORT")
	if port == "" {
		port = "8880"
	}

	log.Printf("Starting api-snowdepth on port %s.\n", port)

	http.Handle("/api/graphql/playground", handler.Playground("GraphQL playground", "/api/graphql"))
	http.Handle("/api/graphql", handler.GraphQL(gql.NewExecutableSchema(gql.Config{Resolvers: &gql.Resolver{}})))

	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS()(r)))
}
