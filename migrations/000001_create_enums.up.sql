CREATE TYPE user_role AS ENUM ('customer', 'owner', 'manager', 'admin');
CREATE TYPE booking_status AS ENUM ('pending', 'confirmed', 'cancelled', 'completed', 'no_show');
CREATE TYPE location_type AS ENUM ('window', 'vip', 'regular', 'outdoor');
CREATE TYPE cuisine_type AS ENUM (
    'Italian', 'Chinese', 'Mexican', 'Japanese', 'Indian', 
    'French', 'Kazakh', 'Turkish', 'Thai', 'American', 
    'Korean', 'Cafe', 'Bar', 'Fast Food', 'Vegetarian', 'Other'
);