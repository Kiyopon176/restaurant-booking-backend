CREATE TYPE transaction_type AS ENUM ('deposit', 'withdraw', 'booking_charge', 'refund', 'payment_to_restaurant');
CREATE TYPE payment_method AS ENUM ('wallet', 'halyk', 'kaspi');
CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed', 'refunded');

-- Wallets table
CREATE TABLE wallets (
                         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                         user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                         balance INTEGER NOT NULL DEFAULT 0 CHECK (balance >= 0),
                         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                         updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);

-- Wallet transactions table
CREATE TABLE wallet_transactions (
                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                     wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
                                     amount INTEGER NOT NULL,
                                     type transaction_type NOT NULL,
                                     description TEXT,
                                     booking_id UUID REFERENCES bookings(id) ON DELETE SET NULL,
                                     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_transactions_booking_id ON wallet_transactions(booking_id);
CREATE INDEX idx_wallet_transactions_created_at ON wallet_transactions(created_at);

-- Payments table
CREATE TABLE payments (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          booking_id UUID REFERENCES bookings(id) ON DELETE SET NULL,
                          amount INTEGER NOT NULL,
                          payment_method payment_method NOT NULL,
                          payment_status payment_status NOT NULL DEFAULT 'pending',
                          external_payment_id VARCHAR(255),
                          external_payment_url TEXT,
                          error_message TEXT,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_booking_id ON payments(booking_id);
CREATE INDEX idx_payments_status ON payments(payment_status);
CREATE INDEX idx_payments_external_id ON payments(external_payment_id);