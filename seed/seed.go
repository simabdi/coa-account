package seed

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"github.com/simabdi/coa-account/model"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

//go:embed coa.xlsx
var coaExcel []byte

type COARow struct {
	Code  string
	Name  string
	Type  string
	Level int
}

// LoadCOAFromExcelBytes reads COA from embedded Excel file
func LoadCOAFromExcelBytes(data []byte) ([]COARow, error) {
	var result []COARow

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("failed to read Excel: " + err.Error())
	}

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, errors.New("failed to read rows: " + err.Error())
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			continue
		}

		level, _ := f.GetCellValue("Sheet1", fmt.Sprintf("D%d", i+1))
		lvl := 0
		fmt.Sscanf(level, "%d", &lvl)

		result = append(result, COARow{
			Code:  row[0],
			Name:  row[1],
			Type:  row[2],
			Level: lvl,
		})
	}

	return result, nil
}

func SeedCOA(db *gorm.DB) error {
	raws, err := LoadCOAFromExcelBytes(coaExcel)
	if err != nil {
		return err
	}

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
