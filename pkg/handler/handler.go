package handler

import (
	"net/http"
	"os"

	
	"github.com/rs/cors"
	"github.com/99designs/gqlgen/handler"
	gql "github.com/iot-for-tillgenglighet/api-snowdepth/internal/pkg/graphql"

	log "github.com/sirupsen/logrus"
)

func Router() {

	mux := http.NewServeMux()

	port := os.Getenv("SNOWDEPTH_API_PORT")
	if port == "" {
		port = "8880"
	}

	log.Printf("Starting api-snowdepth on port %s.\n", port)

	mux.HandleFunc("/api/graphql/playground", handler.Playground("GraphQL playground", "/api/graphql"))
	mux.HandleFunc("/api/graphql", handler.GraphQL(gql.NewExecutableSchema(gql.Config{Resolvers: &gql.Resolver{}})))

	c := cors.Default().Handler(mux)

	log.Fatal(http.ListenAndServe(":"+port, c))
}
