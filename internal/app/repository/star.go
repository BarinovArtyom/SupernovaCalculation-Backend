package repository

import (
	"fmt"
	"lab1/internal/app/ds"
	"math"
)

func (r *Repository) GetStars() ([]ds.Star, error) {
	var stars []ds.Star
	err := r.db.Find(&stars).Error
	if err != nil {
		return nil, err
	}
	if len(stars) == 0 {
		return nil, fmt.Errorf("Массив пустой")
	}
	return stars, nil
}

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

func (r *Repository) Calculation(scope ds.Scope, calc ds.Calc) (float64, float64) {
	E_SN := (4 * math.Pi * math.Pow(calc.InpDist, 2) * calc.InpTexp * scope.Lambda * scope.DeltaLamb) / (3.0e8 * scope.ZeroPoint) * math.Pow(10, (-0.4*calc.InpMass))
	M_Ni56 := 1.2e-7 * (math.Pow(calc.InpDist, 2) / scope.DeltaLamb) * math.Pow(10, -0.4*calc.InpMass) * math.Pow(scope.Lambda/500.0, 2)
	return E_SN, M_Ni56
}

func (r *Repository) DeleteStar(id int) error {
	result := r.db.Model(&ds.Star{}).Where("id = ?", id).Update("status", "deleted")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("star with id %d not found", id)
	}
	return nil
}
