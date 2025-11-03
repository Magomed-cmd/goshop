package mappers

import (
	"time"

	"github.com/shopspring/decimal"
)

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func decimalToString(value decimal.Decimal) string {
	return value.StringFixed(2)
}
