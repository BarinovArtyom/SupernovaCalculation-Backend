package ds

type Users struct {
	ID        int    `json:"id" gorm:"primary_key"`
	Login     string `gorm:"type:varchar(50); not null" json:"login"`
	Password  string `gorm:"type:varchar(255); not null" json:"password"`
	ModStatus bool   `gorm:"not null" json:"mod_status"`
}
