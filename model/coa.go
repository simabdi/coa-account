package model

import "time"

type Coa struct {
	ID        uint      `gorm:"primaryKey;index"`
	Code      string    `gorm:"type:varchar(20);index"`
	Name      string    `gorm:"type:varchar(100);index"`
	Type      string    `gorm:"type:varchar(50);index"`
	Level     int       `gorm:"type:integer(11)"`
	ParentID  *uint     `gorm:"index"`
	Children  []*Coa    `gorm:"-"`
	CreatedBy uint      `gorm:"index"`
	CreatedAt time.Time `gorm:"<-:create;type:datetime(0)"`
	UpdatedAt time.Time `gorm:"<-:update;type:datetime(0)"`
}
