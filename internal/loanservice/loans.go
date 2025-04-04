package loanservice

import (
	"github.com/Dorji/sberInterview/api/protos/services"
)

type LoanCalcService struct {
	services.UnimplementedLoanServiceServer
}

func NewLoanCalcService() *LoanCalcService {
	return &LoanCalcService{cc}
}

func (ls *LoanCalcService) Execute(ctx context.Context, req *services.LoanRequest) (*services.LoanResponse, error) {

	return &services.LoanResponse{Result: "Hello, " + req.Name}, nil
}

func (ls *LoanCalcService) Cache(context.Context, *emptypb.Empty) (*entities.CacheResult, error) {

}

func (ls *LoanCalcService) calculate() {

}
