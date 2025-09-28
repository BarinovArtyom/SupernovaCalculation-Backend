package ds

import "time"

type Star struct {
	ID            int       `json:"id" gorm:"primary_key"`
	Status        string    `gorm:"type:varchar(20); not null" json:"status"`
	CreationDate  time.Time `gorm:"default:CURRENT_TIMESTAMP; not null" json:"creation_date"`
	FormDate      time.Time `gorm:"default:CURRENT_TIMESTAMP; not null" json:"form_date"`
	FinishDate    time.Time `json:"finish_date"`
	Name          string    `gorm:"type:varchar(255)" json:"name"`
	Constellation string    `gorm:"type:varchar(100)" json:"constellation"`
}
