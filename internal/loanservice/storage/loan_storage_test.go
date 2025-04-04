package storage

import (
	"testing"
	"time"
	"sync"

	"github.com/Dorji/sberInterview/api/protos/entities"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLoanCache(t *testing.T) {
	// Создаем тестовые данные
	now := time.Now()
	testLoanResult := &entities.LoanResult{
		Params: &entities.LoanParams{
			ObjectCost:    5000000,
			InitialPayment: 1000000,
			Months:       240,
		},
		Program: &entities.LoanProgram{
			Salary: true,
		},
		Aggregates: &entities.LoanAggregates{
			Rate:           8.0,
			LoanSum:        4000000,
			MonthlyPayment: 33458,
			Overpayment:    4029920,
			LastPaymentDate: timestamppb.New(now.AddDate(20, 0, 0)),
		},
	}

	baseProgram := &entities.LoanProgram{Base: true}
	militaryProgram := &entities.LoanProgram{Military: true}
	salaryProgram := &entities.LoanProgram{Salary: true}

	t.Run("NewLoanCache creates empty cache", func(t *testing.T) {
		cache := NewLoanCache()
		if cache.Size() != 0 {
			t.Errorf("Expected empty cache, got size %d", cache.Size())
		}
	})

	t.Run("Add and Size", func(t *testing.T) {
		cache := NewLoanCache()
		cache.Add(testLoanResult)

		if cache.Size() != 1 {
			t.Errorf("Expected cache size 1, got %d", cache.Size())
		}
	})

	t.Run("GetAll returns correct data", func(t *testing.T) {
		cache := NewLoanCache()
		cache.Add(testLoanResult)

		result := cache.GetAll()
		if len(result.Results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result.Results))
		}

		assert.Equal(t, result.Results[0].Params.ObjectCost, testLoanResult.Params.ObjectCost)
		assert.Equal(t, result.Results[0].Params.InitialPayment, testLoanResult.Params.InitialPayment)
		assert.Equal(t, result.Results[0].Params.Months, testLoanResult.Params.Months)

		assert.Equal(t, result.Results[0].Program.Salary, testLoanResult.Program.Salary)

		assert.Equal(t, result.Results[0].Aggregates.Rate, testLoanResult.Aggregates.Rate)
		assert.Equal(t, result.Results[0].Aggregates.LoanSum, testLoanResult.Aggregates.LoanSum)
		assert.Equal(t, result.Results[0].Aggregates.MonthlyPayment, testLoanResult.Aggregates.MonthlyPayment)
		assert.Equal(t, result.Results[0].Aggregates.Overpayment, testLoanResult.Aggregates.Overpayment)
		assert.Equal(t, result.Results[0].Aggregates.LastPaymentDate, testLoanResult.Aggregates.LastPaymentDate)
		assert.Equal(t, result.Results[0], testLoanResult)

	})

	t.Run("Clear works correctly", func(t *testing.T) {
		cache := NewLoanCache()
		cache.Add(testLoanResult)
		cache.Clear()

		if cache.Size() != 0 {
			t.Errorf("Expected empty cache after Clear, got size %d", cache.Size())
		}
	})

	t.Run("GetByProgram filters correctly", func(t *testing.T) {
		cache := NewLoanCache()
		
		// Добавляем разные программы
		cache.Add(&entities.LoanResult{Program: baseProgram})
		cache.Add(&entities.LoanResult{Program: militaryProgram})
		cache.Add(&entities.LoanResult{Program: salaryProgram})
		cache.Add(&entities.LoanResult{Program: salaryProgram})

		tests := []struct {
			name     string
			program  *entities.LoanProgram
			expected int
		}{
			{"Base program", baseProgram, 1},
			{"Military program", militaryProgram, 1},
			{"Salary program", salaryProgram, 2},
			{"Non-existent program", &entities.LoanProgram{}, 0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := cache.GetByProgram(tt.program)
				if len(result.Results) != tt.expected {
					t.Errorf("Expected %d results for %v, got %d", 
						tt.expected, tt.program, len(result.Results))
				}
			})
		}
	})

	t.Run("Concurrent access", func(t *testing.T) {
		cache := NewLoanCache()
		const goroutines = 100
		var wg sync.WaitGroup

		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				cache.Add(testLoanResult)
				cache.Size()
				cache.GetAll()
			}()
		}
		wg.Wait()

		if cache.Size() != goroutines {
			t.Errorf("Expected %d items, got %d", goroutines, cache.Size())
		}
	})
}

func TestDeepCopy(t *testing.T) {
	cache := NewLoanCache()
	original := &entities.LoanResult{
		Params: &entities.LoanParams{
			ObjectCost:    1000000,
			InitialPayment: 200000,
			Months:       120,
		},
		Program: &entities.LoanProgram{
			Base: true,
		},
		Aggregates: &entities.LoanAggregates{
			Rate:           8,
			LoanSum:        800000,
			MonthlyPayment: 9500,
			Overpayment:    340000,
			LastPaymentDate: timestamppb.New(time.Now()),
		},
	}

	cache.Add(original)
	result := cache.GetAll().Results[0]

	// Модифицируем оригинал
	original.Params.ObjectCost = 999999
	original.Program.Base = false
	original.Aggregates.Rate = 0

	// Проверяем, что копия не изменилась
	if result.Params.ObjectCost == 999999 ||
		!result.Program.Base ||
		result.Aggregates.Rate == 0 {
		t.Error("GetAll() returned a reference, expected deep copy")
	}
}