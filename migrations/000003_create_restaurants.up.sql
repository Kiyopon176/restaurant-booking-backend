CREATE TABLE restaurants (
                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                             owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                             name VARCHAR(255) NOT NULL,
                             address TEXT NOT NULL,
                             latitude DECIMAL(10, 8),
                             longitude DECIMAL(11, 8),
                             description TEXT,
                             phone VARCHAR(20) NOT NULL,
                             instagram VARCHAR(255),
                             website VARCHAR(255),
                             cuisine_type cuisine_type NOT NULL,
                             average_price INTEGER NOT NULL,
                             max_combinable_tables INTEGER NOT NULL DEFAULT 3,
                             working_hours JSONB NOT NULL,
                             rating DECIMAL(2,1) DEFAULT 0.0,
                             reviews_count INTEGER DEFAULT 0,
                             is_active BOOLEAN DEFAULT true,
                             created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                             updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_restaurants_owner_id ON restaurants(owner_id);
CREATE INDEX idx_restaurants_cuisine_type ON restaurants(cuisine_type);
CREATE INDEX idx_restaurants_rating ON restaurants(rating);
CREATE INDEX idx_restaurants_is_active ON restaurants(is_active);