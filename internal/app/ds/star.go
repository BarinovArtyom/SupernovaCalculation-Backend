package ds

import "time"

type Star struct {
	ID            int       `gorm:"primary_key" json:"id"`
	Status        string    `gorm:"type:varchar(20); not null" json:"status"`
	CreationDate  time.Time `gorm:"default:CURRENT_TIMESTAMP; not null" json:"creation_date"`
	FormDate      time.Time `gorm:"default:CURRENT_TIMESTAMP; not null" json:"form_date"`
	FinishDate    time.Time `gorm:"default:CURRENT_TIMESTAMP; not null" json:"finish_date"`
	Name          string    `gorm:"type:varchar(255)" json:"name"`
	Constellation string    `gorm:"type:varchar(100)" json:"constellation"`
	ModID         int       `gorm:"type:int" json:"mod_id"`
	UserID        int       `gorm:"type:int" json:"user_id"`
}
