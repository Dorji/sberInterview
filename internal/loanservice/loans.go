package loanservice

import (
	"context"
	"fmt"
	"math"

	"github.com/Dorji/sberInterview/api/protos/entities"
	"github.com/Dorji/sberInterview/api/protos/services"
	"github.com/Dorji/sberInterview/internal/loanservice/storage"
	"google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type LoanCalcService struct {
	services.UnimplementedLoanServiceServer
}

func NewLoanCalcService() *LoanCalcService {
	return &LoanCalcService{}
}

func (ls *LoanCalcService) Execute(ctx context.Context, req *entities.LoanRequest) (*entities.LoanResponse, error) {
	// Параметры из примера
	loanSum := float64(req.ObjectCost - req.InitialPayment) // Сумма кредита
	annualRate, err := storage.GetAnnualRate(req.Program)
	if err != nil {
		return nil, err
	}
	termMonths := req.Months // Срок
	// Расчет платежа
	monthlyPayment := ls.calculateMonthlyPayment(loanSum, annualRate, termMonths)

	// Округляем до копеек (2 знака после запятой)
	monthlyPayment = math.Round(monthlyPayment*100) / 100

	fmt.Printf("Ежемесячный платеж: %.2f руб.\n", monthlyPayment)

	// Расчет переплаты
	totalPayment := monthlyPayment * float64(termMonths)
	overpayment := totalPayment - loanSum
	fmt.Printf("Общая переплата: %.2f руб.\n", overpayment)
	return &entities.LoanResponse{
		Result: &entities.LoanResult{
			Params: &entities.LoanParams{
				ObjectCost:     req.ObjectCost,
				InitialPayment: req.InitialPayment,
				Months:         req.Months,
			},
			Program: req.Program,
			Aggregates: &entities.LoanAggregates{
				Rate:            annualRate,
				LoanSum:         int64(math.Round(loanSum * 100)),
				MonthlyPayment:  int64(math.Round(monthlyPayment * 100)),
				Overpayment:     int64(math.Round(overpayment * 100)),
				LastPaymentDate: &timestamppb.Timestamp{},
			},
		},
	}, nil
}

func (ls *LoanCalcService) Cache(context.Context, *emptypb.Empty) (*entities.CacheResult, error) {
	return &entities.CacheResult{}, nil
}

func (ls *LoanCalcService) calculateMonthlyPayment(loanSum float64, annualRate float64, months int64) float64 {
	// Конвертируем годовую ставку в месячную
	monthlyRate := annualRate / 12

	// Рассчитываем платеж по формуле аннуитета
	payment := loanSum * (monthlyRate * math.Pow(1+monthlyRate, float64(months))) /
		(math.Pow(1+monthlyRate, float64(months)) - 1)

	return payment
}
