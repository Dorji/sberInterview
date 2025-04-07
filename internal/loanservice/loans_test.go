package loanservice

import (
	// "fmt"
	"fmt"
	"testing"

	"errors"
	"math"
	// storage "github.com/Dorji/sberInterview/internal/loanservice/storage"
)

func TestCalculateMonthlyPayment(t *testing.T) {
	ls := &LoanServiceServer{}

	tests := []struct {
		name        string
		loanSum     int64
		annualRate  float64
		months      int64
		wantPayment int64
		wantErr     error
	}{
		// Корректные случаи
		{
			"Standard case 4M loan",
			4_000_000,
			0.08,
			240,
			33_458,
			nil,
		},
		{
			"Short term 1M loan",
			1_000_000,
			0.12,
			12,
			88_849,
			nil,
		},
		{
			"No interest",
			1200000,
			0.00,
			120,
			0,
			errors.New("calculateMonthlyPayment:installment plan"),
			
		},

		// Ошибочные случаи
		{
			"Zero loan sum",
			0,
			0.08,
			240,
			0,
			errors.New("calculateMonthlyPayment:Zero loan sum"),
		},
		{
			"Negative loan sum",
			-1,
			0.08,
			240,
			0,
			errors.New("calculateMonthlyPayment:Zero loan sum"),
		},
		{
			"Zero months",
			1_000_000,
			0.08,
			0,
			0,
			errors.New("calculateMonthlyPayment:Zero months"),
		},
		{
			"Negative months",
			1_000_000,
			0.08,
			-1,
			0,
			errors.New("calculateMonthlyPayment:Zero months"),
		},
		{
			"Negative rate",
			1_000_000,
			-0.08,
			12,
			0,
			errors.New("calculateMonthlyPayment:rate less than 0.00"),
		},
		{
			"Invalid denominator",
			1_000_000,
			-1.0, // Приведет к отрицательному знаменателю
			12,
			0,
			errors.New("calculateMonthlyPayment:rate less than 0.00"),
		},
		
		{
			"Invalid denominator",
			1_000_000,
			-1.0, // Приведет к отрицательному знаменателю
			12,
			0,
			errors.New("calculateMonthlyPayment:rate less than 0.00"),
		},
		{
			"denominator from very small rate",
			1_000_000,
			0.000000000000001, 
			12,
			0,
			errors.New("calculateMonthlyPayment:denominator less than 0"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ls.calculateMonthlyPayment(tt.loanSum, tt.annualRate, tt.months)

			// Проверка ошибки
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("calculateMonthlyPayment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("calculateMonthlyPayment() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Проверка результата
			if got != tt.wantPayment {
				t.Errorf("calculateMonthlyPayment() = %v, want %v", got, tt.wantPayment)
			}
		})
	}
}

func TestCalculateMonthlyPaymentEdgeCases(t *testing.T) {
	ls := &LoanServiceServer{}

	t.Run("Single month payment", func(t *testing.T) {
		got, err := ls.calculateMonthlyPayment(1000, 0.12, 1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expected := int64(math.Ceil(1000 * (1 + 0.12/12)))
		if got != expected {
			t.Errorf("Expected %v, got %v", expected, got)
		}
	})
}


func TestRoundNotNegative(t *testing.T) {
	ls := &LoanServiceServer{}

	tests := []struct {
		name    string
		num     float64
		want    int64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Positive number",
			num:     123.45,
			want:    124,
			wantErr: false,
		},
		{
			name:    "Negative number",
			num:     -123.45,
			want:    0,
			wantErr: true,
			errMsg:  "LoanServiceServer: num is negative",
		},
		{
			name:    "Negative number 2",
			num:     -0.45,
			want:    0,
			wantErr: true,
			errMsg:  "LoanServiceServer: num is negative",
		},
		{
			name:    "NaN",
			num:     math.NaN(),
			want:    0,
			wantErr: true,
			errMsg:  "LoanServiceServer:NAN",
		},
		{
			name:    "Positive infinity",
			num:     math.Inf(1),
			want:    0,
			wantErr: true,
			errMsg:  "LoanServiceServer:round num overflow",
		},
		{
			name:    "Negative infinity",
			num:     math.Inf(-1),
			want:    0,
			wantErr: true,
			errMsg:  "LoanServiceServer: num is negative",
		},
		{
			name:    "Max int64",
			num:     float64(math.MaxInt64),
			want:    0,
			wantErr: true,
			errMsg:  "LoanServiceServer:round num overflow",
		},
		{
			name:    "Overflow int64",
			num:     float64(math.MaxInt64) + 1,
			want:    0,
			wantErr: true,
			errMsg:  "LoanServiceServer:round num overflow",
		},
		{
			name:    "Zero",
			num:     0.0,
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ls.RoundNotNegative(tt.num)
			if (err != nil) != tt.wantErr {
				fmt.Println(tt.num)
				t.Errorf("RoundNotNegative() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("RoundNotNegative() error message = %v, want %v", err.Error(), tt.errMsg)
			}
			if got != tt.want {
				fmt.Println(tt.num)
				t.Errorf("RoundNotNegative() = %v, want %v", got, tt.want)
			}
		})
	}
}

package loanservice_test

import (
	"context"
	"testing"

	"github.com/Dorji/sberInterview/api/protos/entities"
	"github.com/Dorji/sberInterview/api/protos/services"
	"github.com/Dorji/sberInterview/internal/loanservice"
	"github.com/Dorji/sberInterview/internal/loanservice/storage"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestLoanService_Execute(t *testing.T) {
	tests := []struct {
		name        string
		request     *entities.LoanRequest
		wantErr     bool
		expectedErr codes.Code
	}{
		{
			name: "Valid request",
			request: &entities.LoanRequest{
				ObjectCost:     1_000_000,
				InitialPayment: 200_000,
				Months:        12,
				Program:       "standard",
			},
			wantErr: false,
		},
		{
			name: "Initial payment too low",
			request: &entities.LoanRequest{
				ObjectCost:     1_000_000,
				InitialPayment: 50_000, // Меньше минимального
				Months:        12,
				Program:       "standard",
			},
			wantErr:     true,
			expectedErr: codes.InvalidArgument,
		},
		{
			name: "Invalid program",
			request: &entities.LoanRequest{
				ObjectCost:     1_000_000,
				InitialPayment: 200_000,
				Months:        12,
				Program:       "invalid_program",
			},
			wantErr:     true,
			expectedErr: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := loanservice.NewLoanService(storage.NewLoanCache())
			assert.NoError(t, err)

			resp, err := service.Execute(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if st, ok := status.FromError(err); ok {
					assert.Equal(t, tt.expectedErr, st.Code())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.request.Program, resp.Program)
				assert.Equal(t, tt.request.ObjectCost, resp.Params.ObjectCost)
				assert.True(t, resp.Aggregates.MonthlyPayment > 0)
			}
		})
	}
}

func TestLoanService_Cache(t *testing.T) {
	t.Run("Empty cache", func(t *testing.T) {
		service, err := loanservice.NewLoanService(storage.NewLoanCache())
		assert.NoError(t, err)

		_, err = service.Cache(context.Background(), &emptypb.Empty{})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("With items in cache", func(t *testing.T) {
		cache := storage.NewLoanCache()
		service, err := loanservice.NewLoanService(cache)
		assert.NoError(t, err)

		// Добавляем тестовые данные
		_, err = service.Execute(context.Background(), &entities.LoanRequest{
			ObjectCost:     1_000_000,
			InitialPayment: 200_000,
			Months:        12,
			Program:       "standard",
		})
		assert.NoError(t, err)

		resp, err := service.Cache(context.Background(), &emptypb.Empty{})
		assert.NoError(t, err)
		assert.Len(t, resp.Results, 1)
	})
}