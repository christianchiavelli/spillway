package ingest_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"slices"
	"testing"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/app/ingest"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

func month(y int, m time.Month) time.Time { return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC) }

var (
	windSep   = ingest.File{Key: "wind/2026_09", Source: curtailment.Wind, Month: month(2026, 9), ETag: "a"}
	windOct   = ingest.File{Key: "wind/2026_10", Source: curtailment.Wind, Month: month(2026, 10), ETag: "b"}
	solarOct  = ingest.File{Key: "solar/2026_10", Source: curtailment.Solar, Month: month(2026, 10), ETag: "c"}
	windOld   = ingest.File{Key: "wind/2021_10", Source: curtailment.Wind, Month: month(2021, 10), ETag: "d"}
	quietLogs = slog.New(slog.NewTextHandler(io.Discard, nil))
)

func TestRunLoadsWhatIsNewOrChangedLatestFirst(t *testing.T) {
	catalog := &fakeCatalog{files: []ingest.File{windOld, windSep, solarOct, windOct}}
	store := &fakeStore{versions: map[string]string{
		windSep.Key: windSep.ETag, // loaded, unchanged
		windOct.Key: "an older one",
	}}

	report, err := ingest.New(catalog, store, quietLogs).Run(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if report != (ingest.Report{Files: 4, Loaded: 3}) {
		t.Errorf("report = %+v", report)
	}
	if want := []string{solarOct.Key, windOct.Key, windOld.Key}; !slices.Equal(store.replaced, want) {
		t.Errorf("replaced %v, want %v", store.replaced, want)
	}
	if got := store.months[windOct.Key]; got.Records != 1 || got.Source != curtailment.Wind || len(got.SubsystemDays) != 1 {
		t.Errorf("the month stored for %s = %+v", windOct.Key, got)
	}
	if store.versions[windOct.Key] != windOct.ETag {
		t.Error("the new ETag was not recorded")
	}
}

func TestRunKeepsGoingPastAFileThatFails(t *testing.T) {
	catalog := &fakeCatalog{files: []ingest.File{windOct, solarOct, windSep}, broken: windOct.Key}
	store := &fakeStore{versions: map[string]string{}}

	report, err := ingest.New(catalog, store, quietLogs).Run(t.Context())

	if !errors.Is(err, curtailment.ErrBadRecord) {
		t.Errorf("Run() error = %v, want the failure joined in", err)
	}
	if report != (ingest.Report{Files: 3, Loaded: 2, Failed: 1}) {
		t.Errorf("report = %+v", report)
	}
	if _, ok := store.versions[windOct.Key]; ok {
		t.Error("the broken file was recorded as loaded, so it would never be read again")
	}
}

func TestRunStopsWhenTheListFails(t *testing.T) {
	store := &fakeStore{}
	_, err := ingest.New(&fakeCatalog{listErr: errors.New("connection refused")}, store, quietLogs).Run(t.Context())
	if err == nil || len(store.replaced) != 0 {
		t.Errorf("Run() = %v, replaced %v", err, store.replaced)
	}
}

type fakeCatalog struct {
	files   []ingest.File
	broken  string
	listErr error
}

func (c *fakeCatalog) List(context.Context) ([]ingest.File, error) {
	return slices.Clone(c.files), c.listErr
}

// Read gives each file one record in its month, or a bad one for the broken file.
func (c *fakeCatalog) Read(_ context.Context, f ingest.File, add func(curtailment.Record) error) error {
	r := curtailment.Record{
		Plant: "CJU_A", Subsystem: curtailment.Northeast, GeneratedMW: 10, CurtailedMW: 2,
		At: time.Date(f.Month.Year(), f.Month.Month(), 2, 13, 0, 0, 0, curtailment.Brasilia),
	}
	if f.Key == c.broken {
		r.Subsystem = "SUL"
	}
	return add(r)
}

type fakeStore struct {
	versions map[string]string
	replaced []string
	months   map[string]curtailment.Month
}

func (s *fakeStore) Versions(context.Context) (map[string]string, error) {
	return s.versions, nil
}

func (s *fakeStore) Replace(_ context.Context, f ingest.File, m curtailment.Month) error {
	if s.months == nil {
		s.months = map[string]curtailment.Month{}
	}
	s.replaced = append(s.replaced, f.Key)
	s.months[f.Key] = m
	s.versions[f.Key] = f.ETag
	return nil
}
