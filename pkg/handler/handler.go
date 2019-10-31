package handler

import (
	"encoding/json"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/database"
	"github.com/iot-for-tillgenglighet/api-snowdepth/pkg/models"
)

func Router() {

	router := mux.NewRouter()

	router.HandleFunc("/api/snowdepth/{sensor}", handleSnowdepthRequest).Methods("GET")

	port := os.Getenv("SNOWDEPTH_API_PORT")
	if port == "" {
		port = "8880"
	}

	log.Printf("Starting api-snowdepth on port %s.\n", port)

	err := http.ListenAndServe(":"+port, handlers.LoggingHandler(os.Stdout, router))
	if err != nil {
		log.Print(err)
	}

}

func handleSnowdepthRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sensor := vars["sensor"]

	depth := &models.Snowdepth{}
	database.GetDB().Limit(1).Table("snowdepths").Where("device = ?", sensor).Order("timestamp desc").Find(depth)

	if depth.ID == 0 {
		http.Error(w, "No snowdepth reported for that device", http.StatusNotFound)
		return
	}

	gurka, err := json.MarshalIndent(depth, "", " ")
	if err != nil {

		http.Error(w, "Marshal problem: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(gurka)

}
