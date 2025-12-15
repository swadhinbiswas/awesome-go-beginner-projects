package functionality

import (
  "github.com/google/uuid"
	"gorm.io/gorm"
)

type Profile struct{
  ID string `gorm:"type:text;primaryKey"`
  UserID string `gorm:"type:text;not null;uniqueIndex"`
  Name string
  Bio string
  AvaterUrl string
  Projects []string `gorm:"-:all"`
  numProjects int `gorm:"column:num_projects"`
  Followers int
  Following int


}


func (p *Profile) BeforeCreate(tx *gorm.DB)(err error){
  p.ID =uuid.New().String()
  return nil
}
