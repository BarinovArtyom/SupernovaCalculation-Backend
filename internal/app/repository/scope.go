package repository

import (
	"errors"
	"fmt"
	"time"

	"lab1/internal/app/ds"

	"gorm.io/gorm"
)

func (r *Repository) GetScopes() ([]ds.Scope, error) {
	var scopes []ds.Scope
	err := r.db.Find(&scopes).Error
	if err != nil {
		return nil, err
	}
	if len(scopes) == 0 {
		return nil, fmt.Errorf("Массив пустой")
	}
	return scopes, nil
}

func (r *Repository) GetScopeByID(id int) (ds.Scope, error) {
	var scope ds.Scope
	err := r.db.First(&scope, id).Error
	if err != nil {
		return ds.Scope{}, err
	}
	return scope, nil
}

func (r *Repository) GetScopesByTitle(title string) ([]ds.Scope, error) {
	var scopes []ds.Scope
	err := r.db.Where("name ILIKE ?", "%"+title+"%").Find(&scopes).Error
	if err != nil {
		return nil, err
	}
	return scopes, nil
}

func (r *Repository) GetActiveStar() (int, error) {
	var star ds.Star
	err := r.db.Where("status = ?", "active").First(&star).Error
	if err != nil {
		return 0, err
	}
	return star.ID, nil
}

func (r *Repository) CalcsInStar(starID int) (int, error) {
	var count int64
	err := r.db.Model(&ds.Calc{}).Where("star_id = ?", starID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *Repository) AddToStar(scopeID int) error {
	var star ds.Star
	result := r.db.Where("status = ?", "active").First(&star)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			star = ds.Star{
				Status:        "active",
				CreationDate:  time.Now(),
				FormDate:      time.Now(),
				Name:          "New Star",
				Constellation: "Unknown",
			}

			if err := r.db.Create(&star).Error; err != nil {
				return fmt.Errorf("failed to create star: %w", err)
			}
		} else {
			return fmt.Errorf("failed to find star: %w", result.Error)
		}
	}

	var existingCalc ds.Calc
	result = r.db.Where("star_id = ? AND scope_id = ?", star.ID, scopeID).First(&existingCalc)

	if result.Error == nil {
		return nil
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing calc record: %w", result.Error)
	}

	calc := ds.Calc{
		ScopeID: scopeID,
		StarID:  star.ID,
	}

	if err := r.db.Create(&calc).Error; err != nil {
		return fmt.Errorf("failed to create calc record: %w", err)
	}

	return nil
}
