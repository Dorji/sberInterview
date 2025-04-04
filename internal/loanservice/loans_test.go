package loanservice
import (
	"math"
	"testing"
)

func TestCalculateMonthlyPayment(t *testing.T) {
	tests := []struct {
		name      string
		loanSum   float64
		annualRate float64
		months    int
		want      float64
		wantErr   bool
	}{
		{
			name:      "basic case from example",
			loanSum:   4000000,
			annualRate: 0.08,
			months:    240,
			want:      33458.00,
			wantErr:   false,
		},
		{
			name:      "zero interest rate",
			loanSum:   1000000,
			annualRate: 0.00,
			months:    120,
			want:      1000000 / 120,
			wantErr:   false,
		},
		{
			name:      "short term loan",
			loanSum:   500000,
			annualRate: 0.12,
			months:    12,
			want:      44424.43,
			wantErr:   false,
		},
		{
			name:      "invalid loan sum",
			loanSum:   -100000,
			annualRate: 0.08,
			months:    60,
			want:      0,
			wantErr:   true,
		},
		{
			name:      "invalid rate",
			loanSum:   100000,
			annualRate: -0.05,
			months:    60,
			want:      0,
			wantErr:   true,
		},
		{
			name:      "invalid term",
			loanSum:   100000,
			annualRate: 0.08,
			months:    -10,
			want:      0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := (&LoanCalcService{}).calculateMonthlyPayment(tt.loanSum, tt.annualRate, tt.months)
			
			if !tt.wantErr {
				// Проверяем что результат совпадает с ожидаемым с учетом погрешности округления
				if math.Abs(got-tt.want) > 0.01 {
					t.Errorf("calculateMonthlyPayment() = %v, want %v", got, tt.want)
				}
			} else {
				// Для ошибочных кейсов проверяем что возвращается 0
				if got != 0 {
					t.Errorf("calculateMonthlyPayment() = %v, want 0 for error case", got)
				}
			}
		})
	}
}