package seed

import (
	"github.com/simabdi/coa-account/model"
	"gorm.io/gorm"
)

func SeedCOA(db *gorm.DB) error {
	raws, err := LoadCOAFromExcel("seed/coa.xlsx")
	if err != nil {
		return err
	}

	db.Exec("DELETE FROM coas")

	var insertedCoas []model.Coa
	for _, r := range raws {
		insertedCoas = append(insertedCoas, model.Coa{
			Code:  r.Code,
			Name:  r.Name,
			Type:  r.Type,
			Level: r.Level,
		})
	}

	if err := db.Create(&insertedCoas).Error; err != nil {
		return err
	}

	coaMap := map[string]*model.Coa{}
	for i := range insertedCoas {
		coa := &insertedCoas[i]
		coaMap[coa.Code] = coa
	}

	for i := range insertedCoas {
		current := &insertedCoas[i]
		if current.Level == 1 {
			continue
		}

		for j := i - 1; j >= 0; j-- {
			candidate := insertedCoas[j]
			if candidate.Level == current.Level-1 {
				current.ParentID = &candidate.ID
				break
			}
		}
	}

	for _, coa := range insertedCoas {
		if err := db.Model(&model.Coa{}).
			Where("id = ?", coa.ID).
			Update("parent_id", coa.ParentID).Error; err != nil {
			return err
		}
	}

	return nil
}
