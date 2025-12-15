package functionality

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct{
  ID   string  `gorm:"type:text;primaryKey"`
  Username string
  Email string
  Password string
  Profile Profile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

}


type Login struct{
  username string
  passowrd string
}



func (u *User) BeforeCreate(tx *gorm.DB) (err error){
  u.ID=uuid.New().String()
  hased_password,err:=PassHash(u.Password)
  if err!=nil{
    return err
  }
  u.Password=string(hased_password)
  return nil
}


func (u *User) AfterCreate(tx *gorm.DB)(err error){
  profile:=Profile{
    UserID:u.ID,
    Bio:"",
    AvaterUrl:"",
    numProjects:0,
    Followers:0,
    Following:0,
  }
  result:=tx.Create(&profile)
  return result.Error
}
