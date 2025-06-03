package migrate

import (
	"github.com/simabdi/coa-account/model"
	"github.com/simabdi/coa-account/seed"
	"gorm.io/gorm"
)

func Migration(db *gorm.DB) error {
	status := db.Migrator().HasTable(&model.Coa{})
	if status == false {
		err := db.AutoMigrate(&model.Coa{})
		if err != nil {
			return err
		}

		err = seed.SeedCOA(db)
		if err != nil {
			return err
		}
	}
	return nil
}
