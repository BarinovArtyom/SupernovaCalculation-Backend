package repository

import (
	"errors"
	"fmt"
	"lab1/internal/app/ds"
	"lab1/internal/app/role"
	"math"
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
	star.FormDate = time.Now()

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

func (r *Repository) CalculateStar(starID string) (*[]ds.Calc, error) {
	var star ds.Star
	if err := r.db.First(&star, starID).Error; err != nil {
		return nil, fmt.Errorf("cannot get star: %v", err)
	}

	var calcs []ds.Calc
	if err := r.db.Where("star_id = ?", starID).Find(&calcs).Error; err != nil {
		return nil, fmt.Errorf("cannot get list of calculations: %v", err)
	}

	for i := range calcs {
		var scope ds.Scope
		if err := r.db.First(&scope, calcs[i].ScopeID).Error; err != nil {
			return nil, fmt.Errorf("cannot get scope for calculation %d: %v", calcs[i].StarID, err)
		}

		E_SN, M_Ni56 := r.EnAndNiCalculation(scope, calcs[i])

		// Используем Updates для обновления конкретных полей
		if err := r.db.Model(&ds.Calc{}).
			Where("star_id = ? AND scope_id = ?", starID, calcs[i].ScopeID).
			Updates(map[string]interface{}{
				"res_en": E_SN,
				"res_ni": M_Ni56,
			}).Error; err != nil {
			return nil, fmt.Errorf("cannot save calculation %d: %v", calcs[i].StarID, err)
		}

		// Обновляем локальную копию для возврата
		calcs[i].ResEn = E_SN
		calcs[i].ResNi = M_Ni56
	}

	return &calcs, nil
}

func (r *Repository) EnAndNiCalculation(scope ds.Scope, calc ds.Calc) (float64, float64) {
	E_SN := (4 * math.Pi * math.Pow(calc.InpDist, 2) * calc.InpTexp * scope.Lambda * scope.DeltaLamb) / (3.0e8 * scope.ZeroPoint) * math.Pow(10, (-0.4*calc.InpMass))
	M_Ni56 := 1.2e-7 * (math.Pow(calc.InpDist, 2) / scope.DeltaLamb) * math.Pow(10, -0.4*calc.InpMass) * math.Pow(scope.Lambda/500.0, 2)
	return E_SN, M_Ni56
}

func (r *Repository) SetStarStatus(starID string, status string, modID int) error {
	var star ds.Star
	if err := r.db.First(&star, starID).Error; err != nil {
		return err
	}

	star.Status = status
	star.ModID = modID
	star.FinishDate = time.Now()

	return r.db.Save(&star).Error
}
