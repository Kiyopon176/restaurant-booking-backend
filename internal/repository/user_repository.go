package repository

import (
	"restaurant-booking/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

}

	var user domain.User
		return nil, err
	}
	return &user, nil
}

	var user domain.User
		return nil, err
	}
	return &user, nil
}

	var user domain.User
		return nil, err
	}
	return &user, nil
}

}

}
