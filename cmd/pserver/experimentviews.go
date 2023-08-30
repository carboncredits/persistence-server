package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"quantify.earth/bioserver/internal/database"
)

type ExperimentResponseItem struct {
	Name string `json:"name"`
	ID   uint   `json:"id"`
}

type ExperimentTileResponseItem struct {
	Cell         string  `json:"h3"`
	Count uint    `json:"species"`
	Total     float64 `json:"total_area"`
}

func (s *server) getExperiments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var experiments []database.Experiment
	result := s.db.Find(&experiments)
	if result.Error != nil {
		err_str := fmt.Sprintf("Failed query experiments: %v", result.Error)
		http.Error(w, err_str, http.StatusInternalServerError)
		return
	}

	results := make([]interface{}, len(experiments))
	for idx, experiment := range experiments {
		results[idx] = ExperimentResponseItem{
			Name: experiment.Name,
			ID:   experiment.ID,
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

func (s *server) getTilesForExperiment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	experiment_id := ps.ByName("experimentID")

	var tile_list []ExperimentTileResponseItem

	result := s.db.Model(&database.Tile{}).
		Select("tile as cell, count(species) as count, sum(area) as total").
		Where("experiment_id = ?", experiment_id).
		Group("tile").Find(&tile_list)
	if result.Error != nil {
		err_str := fmt.Sprintf("Failed query experiments: %v", result.Error)
		http.Error(w, err_str, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(tile_list)
	if err != nil {
		log.Printf("Failed to encode get retire response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
