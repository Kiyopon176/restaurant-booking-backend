package domain

func Models() []interface{} {
	return []interface{}{
		&User{},
		&Restaurant{},
		&Table{},
		&Booking{},
		&Review{},
		&RestaurantManager{},
		&RestaurantImage{},
	}
}
