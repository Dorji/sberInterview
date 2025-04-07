package loanservice

import (
	"context"
	"fmt"
	"math"

	"net/http"

	"github.com/Dorji/sberInterview/api/protos/entities"
	"github.com/Dorji/sberInterview/api/protos/services"
	db "github.com/Dorji/sberInterview/internal/db/storage"
	"github.com/Dorji/sberInterview/internal/loanservice/storage"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type LoanServiceServer struct {
	services.UnimplementedLoanServiceServer

	cache *storage.LoanCache
}

func NewLoanService(cache *storage.LoanCache) (*LoanServiceServer, error) {
	res := &LoanServiceServer{cache: cache}
	return res, nil
}

func (ls *LoanServiceServer) Execute(ctx context.Context, req *entities.LoanRequest) (*entities.LoanResult, error) {
	// Параметры из примера
	if float64(req.InitialPayment)/float64(req.ObjectCost) < db.InitialPayment {
		return nil, status.Errorf(http.StatusBadRequest, "the initial payment should be more")
	}
	loanSum := req.ObjectCost - req.InitialPayment // Сумма кредита
	annualRate, err := db.GetAnnualRate(req.Program)
	if err != nil {
		return nil, err
	}
	termMonths := req.Months // Срок
	// Расчет платежа
	monthlyPayment, err := ls.calculateMonthlyPayment(loanSum, annualRate, termMonths)
	if err != nil {
		return nil, err
	}

	// Расчет переплаты
	totalPayment := monthlyPayment * termMonths
	overpayment := totalPayment - loanSum
	res := &entities.LoanResult{
		Params: &entities.LoanParams{
			ObjectCost:     req.ObjectCost,
			InitialPayment: req.InitialPayment,
			Months:         req.Months,
		},
		Program: req.Program,
		Aggregates: &entities.LoanAggregates{
			Rate:            int64(annualRate * 100),
			LoanSum:         loanSum,
			MonthlyPayment:  monthlyPayment,
			Overpayment:     overpayment,
			LastPaymentDate: &timestamppb.Timestamp{},
		},
	}
	ls.cache.Add(res)
	return res, nil
}

func (ls *LoanServiceServer) Cache(context.Context, *emptypb.Empty) (*entities.CacheResult, error) {
	all := ls.cache.GetAll()
	if len(all.Results) == 0 {
		return nil, status.Errorf(http.StatusBadRequest, "empty cache")
	}
	return all, nil
}

func (ls *LoanServiceServer) calculateMonthlyPayment(loanSum int64, annualRate float64, months int64) (int64, error) {
	// Проверка граничных условий
	if loanSum <= 0 {
		return 0, fmt.Errorf("calculateMonthlyPayment:Zero loan sum")
	}
	if months <= 0 {
		return 0, fmt.Errorf("calculateMonthlyPayment:Zero months")
	}

	if loanSum == math.MaxInt64 {
		return 0, fmt.Errorf("calculateMonthlyPayment:loanSum is MaxInt64")
	}

	if months == math.MaxInt64 {
		return 0, fmt.Errorf("calculateMonthlyPayment:months is MaxInt64")
	}
	if annualRate < 0.00 {
		return 0, fmt.Errorf("calculateMonthlyPayment:rate less than 0.00")
	}

	if annualRate == 0.00 {
		return 0, fmt.Errorf("calculateMonthlyPayment:installment plan")
	}

	// Конвертируем годовую ставку в месячную
	monthlyRate := annualRate / 12

	// Защита от деления на ноль и переполнения
	denominator := math.Pow(1+monthlyRate, float64(months)) - 1
	if denominator <= 0 {
		return 0, fmt.Errorf("calculateMonthlyPayment:denominator less than 0")
	}

	// Рассчитываем платеж по формуле аннуитета
	payment := float64(loanSum) * ((monthlyRate * math.Pow(1+monthlyRate, float64(months))) / denominator)

	return ls.RoundNotNegative(payment)
}

func (ls *LoanServiceServer) RoundNotNegative(num float64) (int64, error) {

	if num < 0 {
		return 0, fmt.Errorf("LoanServiceServer: num is negative")
	}

	if math.IsNaN(num) {
		return 0, fmt.Errorf("LoanServiceServer:NAN")
	}
	if math.IsInf(num, 0) {
		if num > 0 {
			return 0, fmt.Errorf("LoanServiceServer:round num overflow")
		}
		return 0, fmt.Errorf("LoanServiceServer:round num overflow")
	}
	if math.IsInf(num, -1) {
		if num < 0 {
			return 0, fmt.Errorf("LoanServiceServer: num is negative")
		}
		return 0, fmt.Errorf("LoanServiceServer: num is negative")
	}

	// Проверка на переполнение int64
	if num >= float64(math.MaxInt64) {
		return 0, fmt.Errorf("LoanServiceServer:round num overflow")
	}

	return int64(math.Ceil(num)), nil
}
