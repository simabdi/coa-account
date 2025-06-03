package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/simabdi/coa-account/model"
	"github.com/simabdi/coa-account/utils"
	"gorm.io/gorm"
	"strconv"
)

type CoaRepository interface {
	WithTx(tx *gorm.DB) CoaRepository
	GetAll(ctx context.Context) ([]*model.Coa, error)
	GetByCode(ctx context.Context, code string) (*model.Coa, error)
	GetByName(ctx context.Context, name string) (*model.Coa, error)
	GetByParentID(ctx context.Context, parentID *uint) ([]*model.Coa, error)
	GetLatestCodeByParentName(ctx context.Context, parentName string) (*model.CoaResponse, error)
	Store(ctx context.Context, data *model.Coa) (*model.Coa, error)
}

type coaRepository struct {
	db *gorm.DB
}

func NewCoaRepository(db *gorm.DB) CoaRepository {
	return &coaRepository{db}
}

func (r *coaRepository) WithTx(tx *gorm.DB) CoaRepository {
	return &coaRepository{db: tx}
}

func (r *coaRepository) GetAll(ctx context.Context) ([]*model.Coa, error) {
	var accounts []model.Coa
	if err := r.db.WithContext(ctx).Find(&accounts).Error; err != nil {
		return nil, err
	}

	return utils.BuildCoaTree(accounts), nil
}

func (r *coaRepository) GetByCode(ctx context.Context, code string) (*model.Coa, error) {
	var user *model.Coa
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *coaRepository) GetByName(ctx context.Context, name string) (*model.Coa, error) {
	var user *model.Coa
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *coaRepository) GetByParentID(ctx context.Context, parentID *uint) ([]*model.Coa, error) {
	var children []*model.Coa

	result := r.db.WithContext(ctx)
	if parentID == nil {
		result = result.Where("parent_id IS NULL")
	} else {
		result = result.Where("parent_id = ?", *parentID)
	}

	err := result.Find(&children).Error
	return children, err
}

func (r *coaRepository) GetLatestCodeByParentName(ctx context.Context, parentName string) (*model.CoaResponse, error) {
	var parent *model.Coa
	err := r.db.WithContext(ctx).Where("name = ?", parentName).First(&parent).Error
	if err != nil {
		return nil, err
	}

	type ParentType struct {
		Type string
	}

	var parentType ParentType
	err = r.db.Table("coas").Select("type").Where("id = ?", parent.ID).Scan(&parentType).Error
	if err != nil {
		return nil, err
	}

	var latestChild *model.Coa
	err = r.db.WithContext(ctx).
		Where("parent_id = ?", parent.ID).
		Order("code DESC").
		First(&latestChild).Error

	var prefixLength, suffixLength int

	switch parent.Level {
	case 2:
		prefixLength = 2
		suffixLength = 4
	case 3:
		prefixLength = 3
		suffixLength = 3
	case 4:
		prefixLength = 4
		suffixLength = 2
	default:
		return nil, fmt.Errorf("unsupported parent level: %d", parent.Level)
	}

	if len(parent.Code) < prefixLength {
		return nil, fmt.Errorf("invalid parent code length")
	}
	prefix := parent.Code[:prefixLength]

	var nextNumber int
	if err == nil {
		if len(latestChild.Code) < prefixLength {
			return nil, fmt.Errorf("invalid child code length")
		}
		suffix := latestChild.Code[prefixLength:]
		currentNum, _ := strconv.Atoi(suffix)
		nextNumber = currentNum + 1
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		nextNumber = 1
	} else {
		return nil, err
	}

	format := fmt.Sprintf("%%s%%0%dd", suffixLength)
	nextCode := fmt.Sprintf(format, prefix, nextNumber)

	return &model.CoaResponse{
		Code:  nextCode,
		Type:  parentType.Type,
		Level: parent.Level + 1,
	}, nil
}

func (r *coaRepository) Store(ctx context.Context, data *model.Coa) (*model.Coa, error) {
	err := r.db.WithContext(ctx).Create(&data).Error
	if err != nil {
		return nil, err
	}

	return data, nil
}
