package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/simabdi/coa-account/model"
	"github.com/simabdi/coa-account/utils"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type CoaRepository interface {
	WithTx(tx *gorm.DB) CoaRepository
	GetAll(ctx context.Context) ([]*model.Coa, error)
	GetByCode(ctx context.Context, code string) (*model.Coa, error)
	GetByName(ctx context.Context, name string) (*model.Coa, error)
	GetByParentID(ctx context.Context, parentID *uint) ([]*model.Coa, error)
	GetLatestCodeByParentName(ctx context.Context, parentName string) (*model.CoaResponse, error)
	GetLatestCodeByParentChild(ctx context.Context, parentName, childName string) (*model.CoaResponse, error)
	GetOrCreateChild(ctx context.Context, parentName, childName string, parentLevel, childLevel int) (*model.CoaResponse, error)
	Store(ctx context.Context, data *model.Coa) (*model.Coa, error)
	BulkStore(ctx context.Context, data []*model.Coa) ([]*model.Coa, error)
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
	err := r.db.WithContext(ctx).Where("LOWER(name) = ?", strings.ToLower(parentName)).First(&parent).Error
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

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

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

	fmt.Println("DB instance coa GetLatestCodeByParentName: ", r.db.Statement.ConnPool)
	return &model.CoaResponse{
		Code:     nextCode,
		Type:     parentType.Type,
		Level:    parent.Level + 1,
		ParentID: parent.ID,
	}, nil
}

func (r *coaRepository) GetLatestCodeByParentChild(ctx context.Context, parentName, childName string) (*model.CoaResponse, error) {
	var parent *model.Coa
	err := r.db.WithContext(ctx).Where("LOWER(name) = ?", strings.ToLower(parentName)).First(&parent).Error
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
		Where("parent_id = ? AND LOWER(name) = ?", parent.ID, strings.ToLower(childName)).
		Order("code DESC").
		First(&latestChild).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found for " + childName)
		}

		return nil, err
	}

	return &model.CoaResponse{
		ID:       latestChild.ID,
		Code:     latestChild.Code,
		Type:     parentType.Type,
		Level:    latestChild.Level,
		ParentID: parent.ID,
	}, nil
}

func (r *coaRepository) GetOrCreateChild(ctx context.Context, parentName, childName string, parentLevel, childLevel int) (*model.CoaResponse, error) {
	var parent *model.Coa

	if err := r.db.WithContext(ctx).
		Where("LOWER(name) = ? AND level = ?", strings.ToLower(parentName), parentLevel).
		First(&parent).Error; err != nil {
		return nil, fmt.Errorf("parent not found: %w", err)
	}

	var child *model.Coa
	if err := r.db.
		Where("LOWER(name) = ? AND level = ? AND parent = ?", strings.ToLower(childName), childLevel, parent.Code).
		First(&child).Error; err == nil {
		return &model.CoaResponse{
			Code:     child.Code,
			Type:     child.Type,
			Level:    child.Level,
			ParentID: *child.ParentID,
		}, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var lastChild *model.Coa
	r.db.WithContext(ctx).
		Where("parent = ? AND level = ?", parent.Code, childLevel).
		Order("code DESC").
		First(&lastChild)

	newCode := ""
	if lastChild.ID == 0 {
		newCode = parent.Code + "01"
	} else {
		lastNumericPart := lastChild.Code[len(parent.Code):]
		var nextNum int
		fmt.Sscanf(lastNumericPart, "%d", &nextNum)
		nextNum++
		newCode = fmt.Sprintf("%s%02d", parent.Code, nextNum)
	}

	type ParentType struct {
		Type string
	}

	var parentType ParentType
	err := r.db.Table("coas").Select("type").Where("id = ?", parent.ID).Scan(&parentType).Error
	if err != nil {
		return nil, err
	}

	newChild := model.Coa{
		Code:     newCode,
		Name:     childName,
		Type:     parentType.Type,
		Level:    childLevel,
		ParentID: &parent.ID,
	}

	if err := r.db.Create(&newChild).Error; err != nil {
		return nil, err
	}

	return &model.CoaResponse{
		Code:     newChild.Code,
		Type:     newChild.Type,
		Level:    newChild.Level,
		ParentID: *newChild.ParentID,
	}, nil
}

func (r *coaRepository) Store(ctx context.Context, data *model.Coa) (*model.Coa, error) {
	err := r.db.WithContext(ctx).Create(&data).Error
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *coaRepository) BulkStore(ctx context.Context, data []*model.Coa) ([]*model.Coa, error) {
	err := r.db.WithContext(ctx).Create(&data).Error
	if err != nil {
		return nil, err
	}

	fmt.Println("DB instance coa BulkStore: ", r.db.Statement.ConnPool)
	return data, nil
}
