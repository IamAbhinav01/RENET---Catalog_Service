package models

type Item struct {
	// using struct tag for validation
	//GORM instructions separated by ;
	ID           int     `gorm:"primaryKey;column:id" json:"id"`
	Title        string  `gorm:"column:title;not null" json:"title"`
	Genres       string  `gorm:"column:genres;not null" json:"genres"`
	PrimaryGenre string  `gorm:"column:primary_genre;not null" json:"primary_genre"`
	PosterURL    *string `gorm:"column:poster_url" json:"poster_url"`
	Plot         *string `gorm:"column:plot" json:"plot"`
	// * > if i get an response where no poster or plot then if normal string then "" , but if * then it returns nil.

}

type Interaction struct {
	ID        int     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int     `gorm:"column:user_id;not null" json:"user_id"`
	ItemID    int     `gorm:"column:item_id;not null" json:"item_id"`
	Rating    float32 `gorm:"column:rating;not null" json:"rating"`
	EventType string  `gorm:"column:event_type;default:'rating'" json:"event_type"`
}

type CreateInteractionRequest struct {
	ItemID    int     `json:"item_id" binding:"required"`
	Rating    float32 `json:"rating" binding:"required"`
	EventType string  `json:"event_type"`
}

type OMDbResponse struct {
	Title    string `json:"Title"`
	Year     string `json:"Year"`
	Plot     string `json:"Plot"`
	Poster   string `json:"Poster"`
	Response string `json:"Response"`
	Error    string `json:"Error"`
}