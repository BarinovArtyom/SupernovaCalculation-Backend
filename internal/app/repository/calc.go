package repository

import (
	"fmt"

	"lab1/internal/app/ds"
)

func (r *Repository) GetCalcs() ([]ds.Calc, error) {
	var calcs []ds.Calc
	err := r.db.Find(&calcs).Error
	if err != nil {
		return nil, err
	}
	if len(calcs) == 0 {
		return nil, fmt.Errorf("Массив пустой")
	}
	return calcs, nil
}

func (r *Repository) GetCalcByID(scopeID int, starID int) (ds.Calc, error) {
	var calc ds.Calc
	err := r.db.Where("scope_id = ? AND star_id = ?", scopeID, starID).First(&calc).Error
	if err != nil {
		return ds.Calc{}, err
	}
	return calc, nil
}
