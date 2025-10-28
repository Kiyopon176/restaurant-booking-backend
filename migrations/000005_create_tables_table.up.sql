CREATE TABLE tables (
                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
                        table_number VARCHAR(50) NOT NULL,
                        min_capacity INTEGER NOT NULL,
                        max_capacity INTEGER NOT NULL,
                        location_type location_type NOT NULL DEFAULT 'regular',
                        x_position INTEGER,
                        y_position INTEGER,
                        is_active BOOLEAN DEFAULT true,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                        CONSTRAINT unique_restaurant_table_number UNIQUE(restaurant_id, table_number)
);

CREATE INDEX idx_tables_restaurant_id ON tables(restaurant_id);
CREATE INDEX idx_tables_is_active ON tables(is_active);