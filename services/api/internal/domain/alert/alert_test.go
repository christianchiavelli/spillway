package alert_test

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/alert"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

func TestNewAcceptsAValidDraft(t *testing.T) {
	id, owner := uuid.NewV7(), uuid.NewV7()
	now := time.Date(2026, 10, 17, 9, 30, 0, 123456789, time.FixedZone("BRT", -3*60*60))

	got, err := alert.New(id, owner, alert.Draft{
		Name:         "  Northeast wind  ",
		Source:       "wind",
		Subsystem:    "NE",
		ThresholdMWh: 500,
	}, now)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	want := alert.Alert{
		ID:           id,
		AccountID:    owner,
		Name:         "Northeast wind",
		Source:       curtailment.Wind,
		Subsystem:    curtailment.Northeast,
		ThresholdMWh: 500,
		CreatedAt:    time.Date(2026, 10, 17, 12, 30, 0, 123456000, time.UTC),
	}
	if got != want {
		t.Errorf("New() = %+v, want %+v", got, want)
	}
}

func TestNewReportsEveryBrokenRule(t *testing.T) {
	tests := []struct {
		name  string
		draft alert.Draft
		want  []domain.Violation
	}{
		{
			name:  "everything missing",
			draft: alert.Draft{},
			want: []domain.Violation{
				{Field: "name", Code: domain.CodeRequired},
				{Field: "source", Code: domain.CodeInvalid},
				{Field: "subsystem", Code: domain.CodeInvalid},
				{Field: "thresholdMwh", Code: domain.CodeOutOfRange},
			},
		},
		{
			name:  "a name of spaces is no name",
			draft: valid(func(d *alert.Draft) { d.Name = " \t " }),
			want:  []domain.Violation{{Field: "name", Code: domain.CodeRequired}},
		},
		{
			name:  "eighty-one letters",
			draft: valid(func(d *alert.Draft) { d.Name = strings.Repeat("ã", 81) }),
			want:  []domain.Violation{{Field: "name", Code: domain.CodeTooLong}},
		},
		{
			name:  "a control character",
			draft: valid(func(d *alert.Draft) { d.Name = "wind\u0007" }),
			want:  []domain.Violation{{Field: "name", Code: domain.CodeInvalid}},
		},
		{
			name:  "subsystem codes are upper case",
			draft: valid(func(d *alert.Draft) { d.Subsystem = "ne" }),
			want:  []domain.Violation{{Field: "subsystem", Code: domain.CodeInvalid}},
		},
		{
			name:  "zero megawatt-hours",
			draft: valid(func(d *alert.Draft) { d.ThresholdMWh = 0 }),
			want:  []domain.Violation{{Field: "thresholdMwh", Code: domain.CodeOutOfRange}},
		},
		{
			name:  "past the largest threshold",
			draft: valid(func(d *alert.Draft) { d.ThresholdMWh = alert.ThresholdMax + 1 }),
			want:  []domain.Violation{{Field: "thresholdMwh", Code: domain.CodeOutOfRange}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := alert.New(uuid.NewV7(), uuid.NewV7(), tt.draft, time.Now())

			var verr *domain.ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("New() error = %v, want a *domain.ValidationError", err)
			}
			if !slices.Equal(verr.Violations, tt.want) {
				t.Errorf("violations = %v, want %v", verr.Violations, tt.want)
			}
		})
	}
}

func TestNewAcceptsTheLimitsThemselves(t *testing.T) {
	for _, d := range []alert.Draft{
		valid(func(d *alert.Draft) { d.Name = strings.Repeat("ã", alert.NameMaxLength) }),
		valid(func(d *alert.Draft) { d.ThresholdMWh = alert.ThresholdMin }),
		valid(func(d *alert.Draft) { d.ThresholdMWh = alert.ThresholdMax }),
	} {
		if _, err := alert.New(uuid.NewV7(), uuid.NewV7(), d, time.Now()); err != nil {
			t.Errorf("New(%+v) error = %v", d, err)
		}
	}
}

func TestDomainErrorsMatchTheirCategory(t *testing.T) {
	if !errors.Is(alert.ErrLimitReached, domain.ErrConflict) {
		t.Error("ErrLimitReached does not match domain.ErrConflict")
	}
	if !errors.Is(alert.ErrNameTaken, domain.ErrConflict) {
		t.Error("ErrNameTaken does not match domain.ErrConflict")
	}
}

func valid(change func(*alert.Draft)) alert.Draft {
	d := alert.Draft{Name: "Wind in the Northeast", Source: "wind", Subsystem: "NE", ThresholdMWh: 500}
	change(&d)
	return d
}
