package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"

	dataframe "github.com/rocketlaunchr/dataframe-go"
	"github.com/rocketlaunchr/dataframe-go/imports"
	"github.com/xitongsys/parquet-go-source/local"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"quantify.earth/bioserver/internal/database"
)

func processSingleSpecies(db *gorm.DB, species_path string, taxa string, experiment database.Experiment) error {
	// Get the species ID from the filename, which should be of the form: res_1234_7.parquet
	basename := path.Base(species_path)
	re := regexp.MustCompile(`res_(\d+)_7.parquet`)
	matches := re.FindStringSubmatch(basename)
	if len(matches) != 2 {
		return fmt.Errorf("failed to find species in %v", basename)
	}
	species_id, err := strconv.Atoi(matches[1])
	if err != nil {
		return fmt.Errorf("failed to parse species id %v: %w", matches[1], err)
	}

	species := database.Species{
		gorm.Model{ID: uint(species_id)},
		taxa,
	}
	err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&species).Error
	if err != nil {
		return err
	}

	fr, err := local.NewLocalFileReader(species_path)
	if err != nil {
		return err
	}
	defer fr.Close()
	ctx := context.Background()
	df, err := imports.LoadFromParquet(ctx, fr)
	if err != nil {
		return err
	}

	// drop tiles with an area of zero
	filterFn := dataframe.FilterDataFrameFn(func(vals map[interface{}]interface{}, row, nRows int) (dataframe.FilterAction, error) {
		if vals["area"].(float64) == 0.0 {
			return dataframe.DROP, nil
		}
		return dataframe.KEEP, nil
	})
	inonzero, err := dataframe.Filter(ctx, df, filterFn)
	if err != nil {
		return err
	}
	nonzero := inonzero.(*dataframe.DataFrame)

	tile_count := nonzero.NRows()
	tiles := make([]database.Tile, tile_count)

	iterator := nonzero.ValuesIterator()
	for {
		row, vals, _ := iterator()
		if row == nil {
			break
		}
		cell := vals["cell"].(string)
		area := vals["area"].(float64)

		tile := database.Tile{
			Tile:         cell,
			Species:      species.ID,
			Area:         area,
			ExperimentID: experiment.ID,
		}
		tiles[*row] = tile
	}
	err = db.CreateInBatches(&tiles, 1000).Error
	if err != nil {
		return err
	}

	return nil
}

func processSpecies(db *gorm.DB, taxa_path string, taxa string, experiment database.Experiment) error {
	contents, err := os.ReadDir(taxa_path)
	if err != nil {
		return err
	}

	for _, subpath := range contents {
		species_path := path.Join(taxa_path, subpath.Name())
		err := processSingleSpecies(db, species_path, taxa, experiment)
		if err != nil {
			return err
		}
	}

	return nil
}

func walkExperiment(db *gorm.DB, experiment_path string) error {
	contents, err := os.ReadDir(experiment_path)
	if err != nil {
		return err
	}

	experiment_name := path.Base(experiment_path)
	experiment := database.Experiment{
		Name: experiment_name,
	}
	err = db.Create(&experiment).Error
	if err != nil {
		return err
	}

	for _, subpath := range contents {
		if !subpath.IsDir() {
			return fmt.Errorf("%s is not a directory", subpath.Name())
		}
		taxaname := subpath.Name()
		taxapath := path.Join(experiment_path, taxaname)
		err := processSpecies(db, taxapath, taxaname, experiment)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Expected experiment folder, containing taxa, containing parquets")
	}

	db := database.DatabaseConnection{
		DSN: os.Getenv("PSERVER_IMPORT_DSN"),
	}
	err := db.InitConnection()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	err = database.Migrate(db.DB)
	if err != nil {
		log.Fatalf("Failed to migrate DB: %v", err)
	}

	log.Printf("Ready to import data from %s", os.Args[1])
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		return walkExperiment(db.DB, os.Args[1])
	})
	if err != nil {
		log.Fatalf("Error walking experiment: %v", err)
	}
}
