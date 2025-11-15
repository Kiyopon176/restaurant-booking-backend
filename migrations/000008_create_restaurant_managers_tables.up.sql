CREATE TABLE restaurant_managers (
                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                     restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
                                     assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                     CONSTRAINT unique_user_restaurant UNIQUE(user_id, restaurant_id)
);

CREATE INDEX idx_restaurant_managers_user_id ON restaurant_managers(user_id);
CREATE INDEX idx_restaurant_managers_restaurant_id ON restaurant_managers(restaurant_id);