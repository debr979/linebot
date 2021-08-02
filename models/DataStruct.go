package models

type DBConnInfo struct {
	USERNAME     string
	USERPASSWORD string
	DBHOST       string
	DBNAME       string
}
type Account struct {
	UID    string `gorm:"column:uid"`
	Total  int    `gorm:"column:total"`
	Credit int    `gorm:"column:credit"`
	Cash   int    `gorm:"column:cash"`
}
type LineCmd struct {
	ID          int    `gorm:"column:id"`
	CMD         string `gorm:"column:cmd"`
	Description string `gorm:"column:description"`
}

