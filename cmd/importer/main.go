package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"quantify.earth/bioserver/internal/database"
)

type CellData struct {
	Cell string `parquet:"name=cell, type=BYTE_ARRAY, repetitiontype=OPTIONAL"`
	Area float64 `parquet:"name=area, type=DOUBLE, repetitiontype=OPTIONAL"`
}

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
	log.Printf("processing species %d", species_id)

	species := database.Species{
		gorm.Model{ID: uint(species_id)},
		taxa,
	}
	err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&species).Error
	if err != nil {
		return err
	}

	log.Printf("opening parquet")
	fr, err := local.NewLocalFileReader(species_path)
	if err != nil {
		return err
	}
	defer fr.Close()

	log.Printf("parsing parquet")
	pr, err := reader.NewParquetReader(fr, new(CellData), 24)
	if err != nil {
		return err
	}
	defer pr.ReadStop()

	rows := int(pr.GetNumRows())
	log.Printf("Row count %d",rows)
	BATCH_SIZE := 1000
	for idx := 0; idx < rows; idx += BATCH_SIZE {
		count := min(BATCH_SIZE, rows - idx)
		buffer := make([]CellData, count)
		err := pr.Read(&buffer)
		if err != nil {
			return err
		}
		non_zero_count := 0
		for _, cell := range(buffer) {
			if cell.Area == 0.0 {
				continue
			}
			non_zero_count += 1
		}
		if non_zero_count == 0 {
			continue
		}

		tiles := make([]database.Tile, non_zero_count)
		tiles_idx := 0
		for row := 0; row < count; row++ {
			cell := buffer[row]
			if cell.Area == 0.0 {
				continue
			}
			tiles[tiles_idx] = database.Tile{
				Tile: cell.Cell,
				Area: cell.Area,
				Species: species.ID,
				ExperimentID: experiment.ID,
			}
			tiles_idx += 1
		}
		err = db.Create(&tiles).Error
		if err != nil {
			return err
		}
	}

	log.Printf("Completed %d", species_id)

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
