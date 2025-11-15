package helper

import "github.com/Kiyopon176/restaurant-booking-backend/internal/domain"

func IsValidPaymentMethod(method domain.PaymentMethod) bool {
	return method == domain.PaymentMethodWallet ||
		method == domain.PaymentMethodHalyk ||
		method == domain.PaymentMethodKaspi
}
