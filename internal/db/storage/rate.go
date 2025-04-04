package storage

import (
	"fmt"

	"github.com/Dorji/sberInterview/api/protos/entities"
)

// подразумевается что они где-то в БД
const (
    InitialPayment     =0.20
	BaseAnnualRate     = 0.10
	MilitaryAnnualRate = 0.09
	SalaryAnnualRate   = 0.08
)

func GetAnnualRate(req *entities.LoanProgram) (float64, error) {
	if req == nil {
		return 0, fmt.Errorf("choose program")
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
		return 0, fmt.Errorf("choose program")
	case selected > 1:
		return 0, fmt.Errorf("choose only 1 program")
	default:
		return rate, nil
	}
}
