CREATE TABLE bookings (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
                          table_id UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
                          user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          booking_date DATE NOT NULL,
                          start_time TIMESTAMP NOT NULL,
                          end_time TIMESTAMP NOT NULL,
                          guests_count INTEGER NOT NULL,
                          status booking_status NOT NULL DEFAULT 'pending',
                          special_note TEXT,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bookings_restaurant_id ON bookings(restaurant_id);
CREATE INDEX idx_bookings_table_id ON bookings(table_id);
CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_booking_date ON bookings(booking_date);
CREATE INDEX idx_bookings_start_time ON bookings(start_time);