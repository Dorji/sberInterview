package storage

import (
	"github.com/Dorji/sberInterview/api/protos/entities"
	"testing"
)

func TestGetAnnualRate(t *testing.T) {
	tests := []struct {
		name     string
		program  *entities.LoanProgram
		wantRate float64
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Base program selected",
			program:  &entities.LoanProgram{Base: true, Military: false, Salary: false},
			wantRate: BaseAnnualRate,
			wantErr:  false,
		},
		{
			name:     "Military program selected",
			program:  &entities.LoanProgram{Base: false, Military: true, Salary: false},
			wantRate: MilitaryAnnualRate,
			wantErr:  false,
		},
		{
			name:     "Salary program selected",
			program:  &entities.LoanProgram{Base: false, Military: false, Salary: true},
			wantRate: SalaryAnnualRate,
			wantErr:  false,
		},
		{
			name:    "No program selected",
			program: &entities.LoanProgram{Base: false, Military: false, Salary: false},
			wantErr: true,
			errMsg:  "rpc error: code = Code(400) desc = choose program",
		},
		{
			name:    "Base and Military selected",
			program: &entities.LoanProgram{Base: true, Military: true, Salary: false},
			wantErr: true,
			errMsg:  "rpc error: code = Code(400) desc = choose only 1 program",
		},
		{
			name:    "Base and Salary selected",
			program: &entities.LoanProgram{Base: true, Military: false, Salary: true},
			wantErr: true,
			errMsg:  "rpc error: code = Code(400) desc = choose only 1 program",
		},
		{
			name:    "Military and Salary selected",
			program: &entities.LoanProgram{Base: false, Military: true, Salary: true},
			wantErr: true,
			errMsg:  "rpc error: code = Code(400) desc = choose only 1 program",
		},
		{
			name:    "All programs selected",
			program: &entities.LoanProgram{Base: true, Military: true, Salary: true},
			wantErr: true,
			errMsg:  "rpc error: code = Code(400) desc = choose only 1 program",
		},
		{
			name:    "NIL statement",
			program: nil,
			wantErr: true,
			errMsg:  "rpc error: code = Code(400) desc = choose program",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRate, err := GetAnnualRate(tt.program)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if gotRate != tt.wantRate {
					t.Errorf("Expected rate %.4f, got %.4f", tt.wantRate, gotRate)
				}
			}
		})
	}
}
