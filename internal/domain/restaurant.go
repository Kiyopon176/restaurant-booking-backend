package domain

import (
	"time"

	"github.com/google/uuid"
)

type Restaurant struct {
	ID                  uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OwnerID             uuid.UUID    `gorm:"type:uuid;not null" json:"owner_id"`
	Name                string       `gorm:"not null" json:"name"`
	Address             string       `gorm:"type:text;not null" json:"address"`
	Latitude            *float64     `json:"latitude,omitempty"`
	Longitude           *float64     `json:"longitude,omitempty"`
	Description         string       `gorm:"type:text" json:"description"`
	Phone               string       `gorm:"not null" json:"phone"`
	Instagram           *string      `json:"instagram,omitempty"`
	Website             *string      `json:"website,omitempty"`
	CuisineType         CuisineType  `gorm:"type:cuisine_type;not null" json:"cuisine_type"`
	AveragePrice        int          `gorm:"not null" json:"average_price"`
	MaxCombinableTables int          `gorm:"not null;default:3" json:"max_combinable_tables"`
	WorkingHours        WorkingHours `gorm:"type:jsonb;not null" json:"working_hours"`
	Rating              float64      `gorm:"type:decimal(2,1);default:0.0" json:"rating"`
	ReviewsCount        int          `gorm:"default:0" json:"reviews_count"`
	IsActive            bool         `gorm:"default:true" json:"is_active"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`

	Owner    *User               `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Images   []RestaurantImage   `gorm:"foreignKey:RestaurantID" json:"images,omitempty"`
	Tables   []Table             `gorm:"foreignKey:RestaurantID" json:"tables,omitempty"`
	Bookings []Booking           `gorm:"foreignKey:RestaurantID" json:"bookings,omitempty"`
	Reviews  []Review            `gorm:"foreignKey:RestaurantID" json:"reviews,omitempty"`
	Managers []RestaurantManager `gorm:"foreignKey:RestaurantID" json:"managers,omitempty"`
}

type CuisineType string

const (
	CuisineTypeItalian    CuisineType = "Italian"
	CuisineTypeChinese    CuisineType = "Chinese"
	CuisineTypeMexican    CuisineType = "Mexican"
	CuisineTypeJapanese   CuisineType = "Japanese"
	CuisineTypeIndian     CuisineType = "Indian"
	CuisineTypeFrench     CuisineType = "French"
	CuisineTypeKazakh     CuisineType = "Kazakh"
	CuisineTypeTurkish    CuisineType = "Turkish"
	CuisineTypeThai       CuisineType = "Thai"
	CuisineTypeAmerican   CuisineType = "American"
	CuisineTypeKorean     CuisineType = "Korean"
	CuisineTypeCafe       CuisineType = "Cafe"
	CuisineTypeBar        CuisineType = "Bar"
	CuisineTypeFastFood   CuisineType = "Fast Food"
	CuisineTypeVegetarian CuisineType = "Vegetarian"
	CuisineTypeOther      CuisineType = "Other"
)

type WorkingHours map[string]DaySchedule

type DaySchedule struct {
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
	IsClosed  bool   `json:"is_closed"`
}

type RestaurantImage struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RestaurantID       uuid.UUID `gorm:"type:uuid;not null" json:"restaurant_id"`
	CloudinaryURL      string    `gorm:"type:text;not null" json:"cloudinary_url"`
	CloudinaryPublicID string    `json:"cloudinary_public_id,omitempty"`
	IsMain             bool      `gorm:"default:false" json:"is_main"`
	CreatedAt          time.Time `json:"created_at"`
}
