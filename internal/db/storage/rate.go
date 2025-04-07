package storage

import (
	"net/http"

	"github.com/Dorji/sberInterview/api/protos/entities"
	"google.golang.org/grpc/status"
)

// подразумевается что они где-то в БД
const (
	InitialPayment     float64 = 0.20
	BaseAnnualRate     float64 = 0.10
	MilitaryAnnualRate float64 = 0.09
	SalaryAnnualRate   float64 = 0.08
)

func GetAnnualRate(req *entities.LoanProgram) (float64, error) {
	if req == nil {
		return 0, status.Errorf(http.StatusBadRequest,"choose program")
	}
	selected := 0
	var rate float64

	if req.Base {
		selected++
		rate = BaseAnnualRate
	}
	if req.Military {
		selected++
		rate = MilitaryAnnualRate
	}
	if req.Salary {
		selected++
		rate = SalaryAnnualRate
	}

	switch {
	case selected == 0:
		return 0, status.Errorf(http.StatusBadRequest, "choose program")
	case selected > 1:
		return 0, status.Errorf(http.StatusBadRequest, "choose only 1 program")
	default:
		return rate, nil
	}
}
