package service

import (
	"context"
	"github.com/simabdi/coa-account/model"
	"github.com/simabdi/coa-account/repository"
)

type CoaService interface {
	FindAll(ctx context.Context) ([]*model.Coa, error)
	FindByCode(ctx context.Context, code string) (*model.Coa, error)
	FindByName(ctx context.Context, name string) (*model.Coa, error)
	FindByParentID(ctx context.Context, parentID *uint) ([]*model.Coa, error)
	FindLatestCodeByParentName(ctx context.Context, parentName string) (*model.CoaResponse, error)
	FindLatestCodeByParentChild(ctx context.Context, parentName, childName string) (*model.CoaResponse, error)
}

type coaService struct {
	repository repository.CoaRepository
}

func NewCoaService(repository repository.CoaRepository) CoaService {
	return &coaService{repository}
}

func (s *coaService) FindAll(ctx context.Context) ([]*model.Coa, error) {
	result, err := s.repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *coaService) FindByCode(ctx context.Context, code string) (*model.Coa, error) {
	result, err := s.repository.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *coaService) FindByName(ctx context.Context, name string) (*model.Coa, error) {
	result, err := s.repository.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *coaService) FindByParentID(ctx context.Context, parentID *uint) ([]*model.Coa, error) {
	return s.repository.GetByParentID(ctx, parentID)
}

func (s *coaService) FindLatestCodeByParentName(ctx context.Context, parentName string) (*model.CoaResponse, error) {
	return s.repository.GetLatestCodeByParentName(ctx, parentName)
}

func (s *coaService) FindLatestCodeByParentChild(ctx context.Context, parentName, childName string) (*model.CoaResponse, error) {
	return s.repository.GetLatestCodeByParentChild(ctx, parentName, childName)
}
