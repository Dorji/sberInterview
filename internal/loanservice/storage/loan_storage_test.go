package storage

import (
	"github.com/Dorji/sberInterview/api/protos/entities"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoanCache(t *testing.T) {
	cache := NewLoanCache()
	baseProgram := &entities.LoanProgram{Base: true}
	militaryProgram := &entities.LoanProgram{Military: true}
	salaryProgram := &entities.LoanProgram{Salary: true}

	params := &entities.LoanParams{
		ObjectCost:     5000000,
		InitialPayment: 1000000,
		Months:         240,
	}

	aggregates := &entities.LoanAggregates{
		Rate:            8.0,
		LoanSum:         4000000,
		MonthlyPayment:  33458,
		Overpayment:     4029920,
		LastPaymentDate: timestamppb.New(time.Now()),
	}

	t.Run("Add and GetAll", func(t *testing.T) {
		cache.Add(params, baseProgram, aggregates)
		cache.Add(params, militaryProgram, aggregates)
		cache.Add(params, salaryProgram, aggregates)

		results := cache.GetAll()
		assert.Equal(t, 3, len(results.Results))
	})

	t.Run("GetByProgram", func(t *testing.T) {
		results := cache.GetByProgram(baseProgram)
		assert.GreaterOrEqual(t, len(results.Results), 1)
	})

	t.Run("Clear", func(t *testing.T) {
		cache.Clear()
		assert.Equal(t, 0, cache.Size())
	})
}
