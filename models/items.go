package models


type Item struct{
	// using struct tag for validation 
	//GORM instructions separated by ;
	ID int `gorm:"primaryKey;column:id" json:"id"`
	Title        string  `gorm:"column:title;not null" json:"title"`
	Genres       string  `gorm:"column:genres;not null" json:"genres"`
	PrimaryGenre string  `gorm:"column:primary_genre;not null" json:"primary_genre"`
	PosterURL    *string `gorm:"column:poster_url" json:"poster_url"`
	Plot         *string `gorm:"column:plot" json:"plot"`
	// * > if i get an response where no poster or plot then if normal string then "" , but if * then it returns nil.
	
}