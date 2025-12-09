package repository

import (
	"errors"
	"fmt"
	"lab1/internal/app/async"
	"lab1/internal/app/ds"
	"lab1/internal/app/role"
	"strconv"
	"strings"
	"time"
)

func (r *Repository) GetStarByID(id int) (ds.Star, error) {
	var star ds.Star
	err := r.db.First(&star, id).Error
	if err != nil {
		return ds.Star{}, err
	}
	return star, nil
}

func (r *Repository) GetScopesByStar(starID int) ([]ds.Scope, error) {
	var scopes []ds.Scope

	err := r.db.
		Joins("JOIN calcs ON calcs.scope_id = scopes.id").
		Where("calcs.star_id = ?", starID).
		Find(&scopes).Error

	if err != nil {
		return nil, err
	}

	return scopes, nil
}

func (r *Repository) GetCalcsByStar(starID int) ([]ds.Calc, error) {
	var calcs []ds.Calc

	err := r.db.
		Where("star_id = ?", starID).
		Find(&calcs).Error

	if err != nil {
		return nil, err
	}

	return calcs, nil
}

func (r *Repository) DeleteStar(id int) error {
	query := "UPDATE stars SET status = 'deleted', form_date = CURRENT_TIMESTAMP WHERE id = $1"

	result := r.db.Exec(query, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("star with id %d not found", id)
	}

	return nil
}

func (r *Repository) GetStars(userRole role.Role, userID int, status string, hasStartDate, hasEndDate bool, startDate, endDate time.Time) (*[]ds.Star, error) {
	var stars []ds.Star

	query := r.db.Model(&ds.Star{})

	query = query.Where("stars.status NOT IN (?, ?)", "active", "deleted")

	if userRole == role.User {
		query = query.Where("stars.user_id = ?", userID)
	}

	if status != "" {
		query = query.Where("stars.status = ?", status)
	}

	if hasStartDate {
		query = query.Where("stars.form_date >= ?", startDate)
	}
	if hasEndDate {
		query = query.Where("stars.form_date <= ?", endDate)
	}

	if err := query.Find(&stars).Error; err != nil {
		return nil, err
	}

	return &stars, nil
}

func (r *Repository) EditStar(id string, name string, constellation string) error {
	var star ds.Star

	if err := r.db.First(&star, id).Error; err != nil {
		return err
	}

	star.Name = name
	star.Constellation = constellation

	return r.db.Save(&star).Error
}

func (r *Repository) FormStar(starID string, creatorID int, calcValues map[string]string) error {
	var star ds.Star

	if err := r.db.First(&star, starID).Error; err != nil {
		return err
	}

	if star.UserID != creatorID {
		return errors.New("attempt to change unowned star")
	}

	if star.Status != "active" {
		return errors.New("Данная заявка не является активной")
	}

	star.Status = "formed"
	now := time.Now()
	star.FormDate = &now

	if err := r.db.Save(&star).Error; err != nil {
		return err
	}

	intStarID, err := strconv.Atoi(starID)
	if err != nil {
		return err
	}

	for scopeIDStr, values := range calcValues {
		scopeID, err := strconv.Atoi(scopeIDStr)
		if err != nil {
			continue
		}

		parts := strings.Fields(values)
		if len(parts) != 3 {
			continue
		}

		inpMass, err1 := strconv.ParseFloat(parts[0], 64)
		inpTexp, err2 := strconv.ParseFloat(parts[1], 64)
		inpDist, err3 := strconv.ParseFloat(parts[2], 64)

		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}
		result := r.db.Model(&ds.Calc{}).
			Where("scope_id = ? AND star_id = ?", scopeID, intStarID).
			Updates(map[string]interface{}{
				"inp_mass": inpMass,
				"inp_texp": inpTexp,
				"inp_dist": inpDist,
			})

		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

func (r *Repository) SetStarStatus(starID string, status string, modID int) error {
	var star ds.Star
	if err := r.db.First(&star, starID).Error; err != nil {
		return err
	}

	star.Status = status
	star.ModID = modID

	if status == "completed" || status == "declined" {
		now := time.Now()
		star.FinishDate = &now
	} else {
		star.FinishDate = nil
	}

	return r.db.Save(&star).Error
}

// GetStarWithCalculations возвращает заявку с расчетами
func (r *Repository) GetStarWithCalculations(starID string) (*ds.Star, []ds.Calc, error) {
	var star ds.Star
	if err := r.db.First(&star, starID).Error; err != nil {
		return nil, nil, fmt.Errorf("cannot get star: %w", err)
	}

	var calcs []ds.Calc
	if err := r.db.Where("star_id = ?", starID).Find(&calcs).Error; err != nil {
		return nil, nil, fmt.Errorf("cannot get calculations: %w", err)
	}

	return &star, calcs, nil
}

// GetCompletedCalculationsCount возвращает количество расчетов с результатами
func (r *Repository) GetCompletedCalculationsCount(starID string) (int, error) {
	var count int64

	err := r.db.Model(&ds.Calc{}).
		Where("star_id = ? AND res_en != 0 AND res_ni != 0", starID).
		Count(&count).Error

	return int(count), err
}

func (r *Repository) SendCalculationsToAsyncService(calcs []ds.Calc, asyncClient *async.AsyncClient, starID string) {
	for _, calc := range calcs {
		var scope ds.Scope
		if err := r.db.First(&scope, calc.ScopeID).Error; err != nil {
			r.logger.Errorf("Cannot get scope for calculation %d: %v", calc.ScopeID, err)
			continue
		}

		if err := asyncClient.SendCalculation(calc, scope, starID); err != nil {
			r.logger.Errorf("Failed to send calculation to async service: %v", err)
		}
	}
}
