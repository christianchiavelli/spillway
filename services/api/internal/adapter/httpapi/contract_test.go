package httpapi

import (
	"errors"
	"strings"
	"testing"

	"github.com/ogen-go/ogen/validate"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/httpapi/oas"
	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/account"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/alert"
)

// The contract and the domain state the same limits in two places: the
// contract for the web app's forms and the edge, the domain for every other
// caller. Each limit is probed on both sides of its edge through the code
// generated from the contract, so changing one without the other fails here.
func TestTheContractAndTheDomainAgreeOnEveryLimit(t *testing.T) {
	probes := []struct {
		name   string
		accept func() error
		refuse func() error
	}{
		{
			name:   "alert name length",
			accept: oas.AlertName(strings.Repeat("ã", alert.NameMaxLength)).Validate,
			refuse: oas.AlertName(strings.Repeat("ã", alert.NameMaxLength+1)).Validate,
		},
		{
			name:   "lowest threshold",
			accept: oas.Threshold(alert.ThresholdMin).Validate,
			refuse: oas.Threshold(alert.ThresholdMin - 1).Validate,
		},
		{
			name:   "highest threshold",
			accept: oas.Threshold(alert.ThresholdMax).Validate,
			refuse: oas.Threshold(alert.ThresholdMax + 1).Validate,
		},
		{
			name:   "list length",
			accept: alertList(alert.MaxPerAccount).Validate,
			refuse: alertList(alert.MaxPerAccount + 1).Validate,
		},
		{
			name:   "email length",
			accept: oas.Email(strings.Repeat("a", account.EmailMaxLength-12) + "@example.com").Validate,
			refuse: oas.Email(strings.Repeat("a", account.EmailMaxLength-11) + "@example.com").Validate,
		},
		{
			name:   "shortest new password",
			accept: newAccount(strings.Repeat("ç", account.PasswordMinLength)).Validate,
			refuse: newAccount(strings.Repeat("ç", account.PasswordMinLength-1)).Validate,
		},
		{
			name:   "longest new password",
			accept: newAccount(strings.Repeat("ç", account.PasswordMaxLength)).Validate,
			refuse: newAccount(strings.Repeat("ç", account.PasswordMaxLength+1)).Validate,
		},
	}
	for _, p := range probes {
		if err := p.accept(); err != nil {
			t.Errorf("%s: the contract refuses what the domain accepts: %v", p.name, err)
		}
		if err := p.refuse(); err == nil {
			t.Errorf("%s: the contract accepts what the domain refuses", p.name)
		}
	}

	for _, s := range []alert.Source{alert.Wind, alert.Solar} {
		if err := oas.Source(s).Validate(); err != nil {
			t.Errorf("source %q is not in the contract", s)
		}
	}
	for _, s := range []alert.Subsystem{alert.North, alert.Northeast, alert.Southeast, alert.South} {
		if err := oas.Subsystem(s).Validate(); err != nil {
			t.Errorf("subsystem %q is not in the contract", s)
		}
	}
	if got, want := len(oas.Subsystem("").AllValues()), 4; got != want {
		t.Errorf("the contract has %d subsystems, the domain %d", got, want)
	}
}

// fieldCode reads one error by its text: ogen reports a number out of range
// with fmt.Errorf. If an upgrade rewords it, this fails, not the form.
func TestFieldCodeReadsWhatTheGeneratedValidatorsReturn(t *testing.T) {
	for _, tt := range []struct {
		err  error
		want domain.Code
	}{
		{oas.Threshold(0).Validate(), domain.CodeOutOfRange},
		{oas.Threshold(alert.ThresholdMax + 1).Validate(), domain.CodeOutOfRange},
		{oas.AlertName("").Validate(), domain.CodeRequired},
		{oas.AlertName(strings.Repeat("a", 81)).Validate(), domain.CodeTooLong},
		{newAccount("short").Validate(), domain.CodeTooShort},
		{oas.Source("coal").Validate(), domain.CodeInvalid},
		{oas.Email("not an address").Validate(), domain.CodeInvalid},
	} {
		if tt.err == nil {
			t.Fatalf("a probe for %s passed validation", tt.want)
		}
		if got := fieldCode(fieldError(tt.err)); got != tt.want {
			t.Errorf("fieldCode(%v) = %s, want %s", tt.err, got, tt.want)
		}
	}
}

func alertList(n int) *oas.AlertList {
	items := make([]oas.Alert, n)
	for i := range items {
		items[i] = oas.Alert{Name: "an alert", Source: oas.SourceWind, Subsystem: oas.SubsystemNE, ThresholdMwh: 1}
	}
	return &oas.AlertList{Items: items}
}

func newAccount(password string) *oas.NewAccount {
	return &oas.NewAccount{Email: "ana@example.com", Password: password}
}

// fieldError returns the error of the one field a *validate.Error holds, as
// the decoder hands it to handleRequestError, or err itself.
func fieldError(err error) error {
	var v *validate.Error
	if errors.As(err, &v) && len(v.Fields) == 1 {
		return v.Fields[0].Error
	}
	return err
}
