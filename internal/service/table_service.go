package service

import (
	"context"
	"errors"
	"fmt"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTableNotFound        = errors.New("table not found")
	ErrInvalidTableNumber   = errors.New("table number cannot be empty")
	ErrInvalidCapacity      = errors.New("min_capacity must be less than or equal to max_capacity")
	ErrDuplicateTableNumber = errors.New("table number already exists for this restaurant")
)

type CreateTableRequest struct {
	TableNumber  string
	MinCapacity  int
	MaxCapacity  int
	LocationType domain.LocationType
	XPosition    *int
	YPosition    *int
}

type UpdateTableRequest struct {
	TableNumber  *string
	MinCapacity  *int
	MaxCapacity  *int
	LocationType *domain.LocationType
	XPosition    *int
	YPosition    *int
	IsActive     *bool
}

type BulkCreateTablesRequest struct {
	Tables []CreateTableRequest
}

type TableService interface {
	CreateTable(ctx context.Context, restaurantID uuid.UUID, ownerID uuid.UUID, req CreateTableRequest) (*domain.Table, error)
	GetTablesByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*domain.Table, error)
	UpdateTable(ctx context.Context, id uuid.UUID, restaurantID uuid.UUID, ownerID uuid.UUID, req UpdateTableRequest) (*domain.Table, error)
	DeleteTable(ctx context.Context, id uuid.UUID, restaurantID uuid.UUID, ownerID uuid.UUID) error
	BulkCreateTables(ctx context.Context, restaurantID uuid.UUID, ownerID uuid.UUID, req BulkCreateTablesRequest) ([]*domain.Table, error)
}

type tableService struct {
	tableRepo      repository.TableRepository
	restaurantRepo repository.RestaurantRepository
	db             *gorm.DB
}

func NewTableService(tableRepo repository.TableRepository, restaurantRepo repository.RestaurantRepository, db *gorm.DB) TableService {
	return &tableService{
		tableRepo:      tableRepo,
		restaurantRepo: restaurantRepo,
		db:             db,
	}
}

func (s *tableService) CreateTable(ctx context.Context, restaurantID uuid.UUID, ownerID uuid.UUID, req CreateTableRequest) (*domain.Table, error) {
	restaurant, err := s.restaurantRepo.GetByID(ctx, restaurantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRestaurantNotFound
		}
		return nil, err
	}

	if restaurant.OwnerID != ownerID {
		return nil, ErrUnauthorized
	}

	if strings.TrimSpace(req.TableNumber) == "" {
		return nil, ErrInvalidTableNumber
	}

	if req.MinCapacity > req.MaxCapacity {
		return nil, ErrInvalidCapacity
	}

	if err := s.checkDuplicateTableNumber(ctx, restaurantID, req.TableNumber, uuid.Nil); err != nil {
		return nil, err
	}

	table := &domain.Table{
		RestaurantID: restaurantID,
		TableNumber:  req.TableNumber,
		MinCapacity:  req.MinCapacity,
		MaxCapacity:  req.MaxCapacity,
		LocationType: req.LocationType,
		XPosition:    req.XPosition,
		YPosition:    req.YPosition,
		IsActive:     true,
	}

	if err := s.tableRepo.Create(ctx, table); err != nil {
		return nil, err
	}

	return table, nil
}

func (s *tableService) GetTablesByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*domain.Table, error) {
	return s.tableRepo.GetByRestaurantID(ctx, restaurantID)
}

func (s *tableService) UpdateTable(ctx context.Context, id uuid.UUID, restaurantID uuid.UUID, ownerID uuid.UUID, req UpdateTableRequest) (*domain.Table, error) {
	// Verify ownership
	restaurant, err := s.restaurantRepo.GetByID(ctx, restaurantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRestaurantNotFound
		}
		return nil, err
	}

	if restaurant.OwnerID != ownerID {
		return nil, ErrUnauthorized
	}

	table, err := s.tableRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTableNotFound
		}
		return nil, err
	}

	if table.RestaurantID != restaurantID {
		return nil, ErrUnauthorized
	}

	if req.TableNumber != nil {
		if strings.TrimSpace(*req.TableNumber) == "" {
			return nil, ErrInvalidTableNumber
		}
		if err := s.checkDuplicateTableNumber(ctx, restaurantID, *req.TableNumber, id); err != nil {
			return nil, err
		}
		table.TableNumber = *req.TableNumber
	}

	if req.MinCapacity != nil {
		table.MinCapacity = *req.MinCapacity
	}

	if req.MaxCapacity != nil {
		table.MaxCapacity = *req.MaxCapacity
	}

	// Validate capacity after updates
	if table.MinCapacity > table.MaxCapacity {
		return nil, ErrInvalidCapacity
	}

	if req.LocationType != nil {
		table.LocationType = *req.LocationType
	}

	if req.XPosition != nil {
		table.XPosition = req.XPosition
	}

	if req.YPosition != nil {
		table.YPosition = req.YPosition
	}

	if req.IsActive != nil {
		table.IsActive = *req.IsActive
	}

	if err := s.tableRepo.Update(ctx, table); err != nil {
		return nil, err
	}

	return table, nil
}

func (s *tableService) DeleteTable(ctx context.Context, id uuid.UUID, restaurantID uuid.UUID, ownerID uuid.UUID) error {
	restaurant, err := s.restaurantRepo.GetByID(ctx, restaurantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRestaurantNotFound
		}
		return err
	}

	if restaurant.OwnerID != ownerID {
		return ErrUnauthorized
	}

	table, err := s.tableRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTableNotFound
		}
		return err
	}

	if table.RestaurantID != restaurantID {
		return ErrUnauthorized
	}

	table.IsActive = false
	return s.tableRepo.Update(ctx, table)
}

func (s *tableService) BulkCreateTables(ctx context.Context, restaurantID uuid.UUID, ownerID uuid.UUID, req BulkCreateTablesRequest) ([]*domain.Table, error) {
	restaurant, err := s.restaurantRepo.GetByID(ctx, restaurantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRestaurantNotFound
		}
		return nil, err
	}

	if restaurant.OwnerID != ownerID {
		return nil, ErrUnauthorized
	}

	tableNumbers := make(map[string]bool)
	for i, tableReq := range req.Tables {
		if strings.TrimSpace(tableReq.TableNumber) == "" {
			return nil, fmt.Errorf("table at index %d: %w", i, ErrInvalidTableNumber)
		}

		if tableReq.MinCapacity > tableReq.MaxCapacity {
			return nil, fmt.Errorf("table at index %d: %w", i, ErrInvalidCapacity)
		}

		if tableNumbers[tableReq.TableNumber] {
			return nil, fmt.Errorf("table at index %d: duplicate table number '%s' in request", i, tableReq.TableNumber)
		}
		tableNumbers[tableReq.TableNumber] = true

		if err := s.checkDuplicateTableNumber(ctx, restaurantID, tableReq.TableNumber, uuid.Nil); err != nil {
			return nil, fmt.Errorf("table at index %d: %w", i, err)
		}
	}

	tables := make([]*domain.Table, len(req.Tables))
	for i, tableReq := range req.Tables {
		table := &domain.Table{
			RestaurantID: restaurantID,
			TableNumber:  tableReq.TableNumber,
			MinCapacity:  tableReq.MinCapacity,
			MaxCapacity:  tableReq.MaxCapacity,
			LocationType: tableReq.LocationType,
			XPosition:    tableReq.XPosition,
			YPosition:    tableReq.YPosition,
			IsActive:     true,
		}

		if err := s.tableRepo.Create(ctx, table); err != nil {
			return nil, fmt.Errorf("failed to create table at index %d: %w", i, err)
		}

		tables[i] = table
	}

	return tables, nil
}

func (s *tableService) checkDuplicateTableNumber(ctx context.Context, restaurantID uuid.UUID, tableNumber string, excludeID uuid.UUID) error {
	var count int64
	query := s.db.WithContext(ctx).
		Model(&domain.Table{}).
		Where("restaurant_id = ? AND table_number = ?", restaurantID, tableNumber)

	if excludeID != uuid.Nil {
		query = query.Where("id != ?", excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return ErrDuplicateTableNumber
	}

	return nil
}
