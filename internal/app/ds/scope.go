package ds

type Scope struct {
	ID          int     `json:"id" gorm:"primary_key"`
	Name        string  `gorm:"type:varchar(255)" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	Status      bool    `gorm:"default:true" json:"status"`
	ImgLink     string  `gorm:"type:varchar(500)" json:"img_link"`
	Filter      string  `gorm:"type:varchar(100)" json:"filter"`
	Lambda      float64 `json:"lambda"`
	DeltaLamb   float64 `json:"delta_lamb"`
	ZeroPoint   float64 `json:"zero_point"`
}
