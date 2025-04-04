package middleware

// import (
// 	"context"
// 	"fmt"

// 	"github.com/Dorji/sberInterview/api/protos/entities"
// 	"github.com/Dorji/sberInterview/api/protos/services"
// 	"github.com/Dorji/sberInterview/internal/loanservice"
// 	"github.com/Dorji/sberInterview/internal/loanservice/storage"
// 	"google.golang.org/protobuf/types/known/emptypb"
// )

// type DecoratorLoanServiceServer struct {
// 	services.UnimplementedLoanServiceServer
// 	service *services.LoanServiceServer
// }

// func NewDecoratorLoanService(cache *storage.LoanCache,srvc services.UnimplementedLoanServiceServer ) (*DecoratorLoanServiceServer, error) {
// 	if cache ==nil{
// 		return nil,fmt.Errorf("NO cache")
// 	}
// 	ls,err:=loanservice.NewLoanService(cache)
// 	if err!=nil{
// 		return nil,err
// 	}
// 	res:=&DecoratorLoanServiceServer{ service: *ls}
// 	return res,nil
// }

// func (dls *DecoratorLoanServiceServer) Execute(ctx context.Context, req *entities.LoanRequest) (*entities.LoanResult, error) {
// 	dls.service.Execute(ctx,req)
// 	return res, nil
// }

// func (dls *DecoratorLoanServiceServer) Cache(context.Context, *emptypb.Empty) (*entities.CacheResult, error) {

// 	return &entities.CacheResult{}, nil
// }

// func (ls *DecoratorLoanServiceServer) calculateMonthlyPayment(loanSum int64, annualRate float64, months int64) (int64, error) {
	
// }

// func (ls *DecoratorLoanServiceServer) RoundNotNegative(num float64) (int64, error) {
// 	return
// }
