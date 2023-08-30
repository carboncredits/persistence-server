package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"quantify.earth/bioserver/internal/database"
)

type SpeciesResponseItem struct {
	ID   uint   `json:"id"`
	Taxa string `json:"taxa"`
}

type APIResponse struct {
	Data []interface{} `json:"data"`
}

func (s *server) getSpecies(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var species []database.Species
	result := s.db.Find(&species)
	if result.Error != nil {
		err_str := fmt.Sprintf("Failed query species: %v", result.Error)
		http.Error(w, err_str, http.StatusInternalServerError)
		return
	}

	results := make([]interface{}, len(species))
	for idx, spec := range species {
		results[idx] = SpeciesResponseItem{
			ID:   spec.ID,
			Taxa: spec.Taxa,
		}
	}

	response := APIResponse{Data: results}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Failed to encode get retire response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
