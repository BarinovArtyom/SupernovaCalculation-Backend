package ds

import "lab1/internal/app/role"

type Users struct {
	ID        int       `json:"id" gorm:"primary_key"`
	Login     string    `gorm:"type:varchar(50); not null" json:"login"`
	Password  string    `gorm:"type:varchar(255); not null" json:"password"`
	ModStatus role.Role `gorm:"not null" json:"mod_status"`
}
