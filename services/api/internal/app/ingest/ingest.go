// Package ingest is the use case of keeping Spillway's copy of the ONS's
// curtailment data in step with what the ONS publishes: one file per source
// and month, which the ONS rewrites whenever it adds a day or revises one.
package ingest

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

// File is one of the ONS's monthly files.
type File struct {
	Key    string // where the ONS keeps it
	Source curtailment.Source
	Month  time.Time // its first day
	ETag   string    // changes whenever the file's content does
}

// Catalog is where the ONS publishes its files.
type Catalog interface {
	// List returns every monthly file of both sources.
	List(ctx context.Context) ([]File, error)
	// Read passes each record of f to add, and stops at the first error.
	Read(ctx context.Context, f File, add func(curtailment.Record) error) error
}

// Store keeps the sums of the files.
type Store interface {
	// Versions returns the ETag of every file loaded, by key.
	Versions(ctx context.Context) (map[string]string, error)
	// Replace swaps whatever is kept of m's source and month for m, and
	// records f's ETag, in one go: a reader sees the month before or after,
	// never half of each.
	Replace(ctx context.Context, f File, m curtailment.Month) error
}

// Report says what a run did.
type Report struct {
	Files  int // the files the ONS lists
	Loaded int // the new and changed ones loaded
	Failed int // the ones that failed, left for the next run
}

// Service runs the use case.
type Service struct {
	catalog Catalog
	store   Store
	logger  *slog.Logger
}

// New returns a Service that logs each file it loads to logger.
func New(catalog Catalog, store Store, logger *slog.Logger) *Service {
	return &Service{catalog: catalog, store: store, logger: logger}
}

// Run loads every file that is new, or has changed since it was loaded. The
// latest months go first, so a fresh start shows the latest figures within
// seconds and fills the history behind them. A file that fails is left for the
// next run and the rest still load; the error joins every failure.
func (s *Service) Run(ctx context.Context) (Report, error) {
	files, err := s.catalog.List(ctx)
	if err != nil {
		return Report{}, fmt.Errorf("list the ONS's files: %w", err)
	}
	versions, err := s.store.Versions(ctx)
	if err != nil {
		return Report{}, fmt.Errorf("read the files loaded: %w", err)
	}

	slices.SortFunc(files, func(a, b File) int {
		return cmp.Or(b.Month.Compare(a.Month), cmp.Compare(a.Source, b.Source))
	})
	report := Report{Files: len(files)}
	var errs []error
	for _, f := range files {
		if versions[f.Key] == f.ETag {
			continue
		}
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			break
		}
		if err := s.load(ctx, f); err != nil {
			report.Failed++
			errs = append(errs, fmt.Errorf("%s: %w", f.Key, err))
			s.logger.WarnContext(ctx, "file not loaded", slog.String("file", f.Key), slog.Any("error", err))
			continue
		}
		report.Loaded++
	}
	return report, errors.Join(errs...)
}

func (s *Service) load(ctx context.Context, f File) error {
	start := time.Now()
	b := curtailment.NewMonthBuilder(f.Source, f.Month)
	if err := s.catalog.Read(ctx, f, b.Add); err != nil {
		return fmt.Errorf("read: %w", err)
	}
	m := b.Month()
	if err := s.store.Replace(ctx, f, m); err != nil {
		return fmt.Errorf("store: %w", err)
	}
	s.logger.InfoContext(ctx, "file loaded",
		slog.String("file", f.Key),
		slog.Int("records", m.Records),
		slog.Int("plants", len(m.Plants)),
		slog.Duration("took", time.Since(start)))
	return nil
}
