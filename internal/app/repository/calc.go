package repository

import (
	"fmt"
	"lab1/internal/app/ds"
)

func (r *Repository) GetEnCalcs() ([]ds.Calc, error) {
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

func (r *Repository) GetEnCalcByID(scopeID int, starID int) (ds.Calc, error) {
	var calc ds.Calc
	err := r.db.Where("scope_id = ? AND star_id = ?", scopeID, starID).First(&calc).Error
	if err != nil {
		return ds.Calc{}, err
	}
	return calc, nil
}

func (r *Repository) DeleteCalcFromStar(starID string, scopeID string) error {
	query := "DELETE FROM calcs WHERE star_id = ? AND scope_id = ?"
	return r.db.Exec(query, starID, scopeID).Error
}

func (r *Repository) EditCalcInStar(starID, scopeID int, inpMass, inpTexp, inpDist float64) error {
	var calc ds.Calc
	if err := r.db.Where("star_id = ? AND scope_id = ?", starID, scopeID).First(&calc).Error; err != nil {
		return fmt.Errorf("calculation not found: %w", err)
	}

	result := r.db.Model(&ds.Calc{}).
		Where("star_id = ? AND scope_id = ?", starID, scopeID).
		Updates(map[string]interface{}{
			"inp_mass": inpMass,
			"inp_texp": inpTexp,
			"inp_dist": inpDist,
			"res_en":   0,
			"res_ni":   0,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update calculation: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows affected - calculation not found")
	}

	return nil
}
