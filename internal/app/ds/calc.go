package ds

type Calc struct {
	ScopeID int     `gorm:"not null" json:"scope_id"`
	StarID  int     `gorm:"not null" json:"star_id"`
	InpMass float64 `json:"inp_mass"`
	InpTexp float64 `json:"inp_texp"`
	InpDist float64 `json:"inp_dist"`
	ResEn   float64 `json:"res_en"`
	ResNi   float64 `json:"res_ni"`
}
