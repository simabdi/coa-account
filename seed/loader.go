package seed

import (
	"github.com/xuri/excelize/v2"
	"strconv"
)

type CoaRaw struct {
	Code  string
	Name  string
	Type  string
	Level int
}

func LoadCOAFromExcel(path string) ([]CoaRaw, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	var coas []CoaRaw
	for idx, row := range rows {
		if idx == 0 {
			continue // skip header
		}
		if len(row) < 4 {
			continue
		}
		level, err := strconv.Atoi(row[3])
		if err != nil {
			continue
		}
		coas = append(coas, CoaRaw{
			Code:  row[0],
			Name:  row[1],
			Type:  row[2],
			Level: level,
		})
	}
	return coas, nil
}
