package utils

import (
	"github.com/simabdi/coa-account/model"
)

func BuildCoaTree(flat []model.Coa) []*model.Coa {
	idMap := make(map[uint]*model.Coa)
	for i := range flat {
		acc := &flat[i]
		idMap[acc.ID] = acc
	}

	var roots []*model.Coa
	for i := range flat {
		acc := &flat[i]
		if acc.ParentID != nil {
			if parent, ok := idMap[*acc.ParentID]; ok {
				parent.Children = append(parent.Children, acc)
			}
		} else {
			roots = append(roots, acc)
		}
	}

	return roots
}
