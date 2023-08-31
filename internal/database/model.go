package database

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type Species struct {
	gorm.Model
	Taxa string `gorm:"type:varchar(32);not null;index"`
}

type Experiment struct {
	gorm.Model
	Name string `gorm:"type:varchar(128);not null;unique"`
}

type Tile struct {
	ID           uint    `gorm:"primaryKey;not null"`
	Tile         string  `gorm:"type:varchar(16);index:idx_tile_tile;index:idx_tile_unique,unique;not null"`
	Species      uint    `gorm:"index:idx_tile_species;index:idx_tile_unique,unique;not null"`
	Area         float64 `gorm:"not null"`
	ExperimentID uint    `gorm:"index:idx_tile_experiment;index:idx_tile_unique,unique;not null"`
	// Center     EWKBGeomPoint `gorm:"column:geom"`
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("Go unexpected nil database connection")
	}

	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202308301428",
			Migrate: func(tx *gorm.DB) error {
				type experiment struct {
					gorm.Model
					Name string `gorm:"type:varchar(128);not null;unique"`
				}
				err := tx.AutoMigrate(&experiment{})
				if err != nil {
					return err
				}
				type species struct {
					gorm.Model
					Taxa string `gorm:"type:varchar(32);not null;index"`
				}
				err = tx.AutoMigrate(&species{})
				if err != nil {
					return err
				}
				type tile struct {
					ID           uint    `gorm:"primaryKey;not null"`
					Tile         string  `gorm:"type:varchar(16);index:idx_tile_tile;index:idx_tile_unique,unique;not null"`
					Species      uint    `gorm:"index:idx_tile_species;index:idx_tile_unique,unique;not null"`
					Area         float64 `gorm:"not null"`
					ExperimentID uint    `gorm:"index:idx_tile_experiment;index:idx_tile_unique,unique;not null"`
				}
				return tx.AutoMigrate(&tile{})
			},
			Rollback: func(tx *gorm.DB) error {
				err := tx.Migrator().DropTable("tiles")
				if err != nil {
					return err
				}
				err = tx.Migrator().DropTable("experiments")
				if err != nil {
					return err
				}
				return tx.Migrator().DropTable("species")
			},
		},
	})

	return m.Migrate()
}
