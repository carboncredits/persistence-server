package main

import (
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"

	"quantify.earth/bioserver/internal/database"
)

type server struct {
	mux *httprouter.Router
	db  *gorm.DB
}

func SetupMyHandlers(db *gorm.DB) server {
	router := httprouter.New()
	server := server{
		mux: router,
		db:  db,
	}

	router.GET("/api/experiments/", server.getExperiments)
	router.GET("/api/experiments/:experimentID/tiles/", server.getTilesForExperiment)
	router.GET("/api/species/", server.getSpecies)

	return server
}

func main() {

	db := database.DatabaseConnection{
		DSN: os.Getenv("PSERVER_DSN"),
	}
	err := db.InitConnection()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	err = database.Migrate(db.DB)
	if err != nil {
		log.Fatalf("Failed to migrate DB: %v", err)
	}

	server := SetupMyHandlers(db.DB)
	http.ListenAndServe(":8080", server.mux)
}
