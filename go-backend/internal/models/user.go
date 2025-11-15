package models

type User struct {
	Email         string  `gorm:"column:email;type:varchar(100);primaryKey"`
	FirstName     string  `gorm:"column:firstName;type:varchar(50)"`
	LastName      string  `gorm:"column:lastName;type:varchar(50)"`
	UserThumbnail *string `gorm:"column:userThumbnail;type:varchar(255)"`
	Location      *string `gorm:"column:location;type:varchar(100)"`
}

func (User) TableName() string {
	return "users_"
}
