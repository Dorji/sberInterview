package storage

import (
	"sync"

	"github.com/Dorji/sberInterview/api/protos/entities"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LoanCache struct {
	mu    sync.RWMutex
	items []*entities.LoanResult
}

func NewLoanCache() *LoanCache {
	return &LoanCache{
		items: make([]*entities.LoanResult, 0),
	}
}

// Add добавляет новый результат расчета в кеш
func (c *LoanCache) Add(entity *entities.LoanResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = append(c.items, entity)
}

// GetAll возвращает все результаты в виде CacheResult
func (c *LoanCache) GetAll() *entities.CacheResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Создаем глубокую копию для безопасности
	results := make([]*entities.LoanResult, len(c.items))
	for i, item := range c.items {
		results[i] = &entities.LoanResult{
			Params: &entities.LoanParams{
				ObjectCost:     item.Params.ObjectCost,
				InitialPayment: item.Params.InitialPayment,
				Months:         item.Params.Months,
			},
			Program: &entities.LoanProgram{
				Salary:   item.Program.Salary,
				Military: item.Program.Military,
				Base:     item.Program.Base,
			},
			Aggregates: &entities.LoanAggregates{
				Rate:            item.Aggregates.Rate,
				LoanSum:         item.Aggregates.LoanSum,
				MonthlyPayment:  item.Aggregates.MonthlyPayment,
				Overpayment:     item.Aggregates.Overpayment,
				LastPaymentDate: timestamppb.New(item.Aggregates.LastPaymentDate.AsTime()),
			},
		}
	}

	return &entities.CacheResult{
		Results: results,
	}
}

// Clear очищает кеш
func (c *LoanCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make([]*entities.LoanResult, 0)
}

// Size возвращает текущее количество элементов в кеше
func (c *LoanCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// GetByProgram возвращает результаты только для указанной программы
func (c *LoanCache) GetByProgram(program *entities.LoanProgram) *entities.CacheResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var filtered []*entities.LoanResult
	for _, item := range c.items {
		if item.Program.Salary == program.Salary &&
			item.Program.Military == program.Military &&
			item.Program.Base == program.Base {
			filtered = append(filtered, item)
		}
	}

	return &entities.CacheResult{
		Results: filtered,
	}
}
