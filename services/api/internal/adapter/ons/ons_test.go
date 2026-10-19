package ons_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/deprecated"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/ons"
	"github.com/christianchiavelli/spillway/services/api/internal/app/ingest"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

const (
	windPrefix  = "dataset/restricao_coff_eolica_tm/"
	solarPrefix = "dataset/restricao_coff_fotovoltaica_tm/"
)

// listing is S3's answer to ListObjectsV2, as the ONS's bucket gives it.
func listing(truncated bool, next string, keys ...string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">`)
	fmt.Fprintf(&b, "<IsTruncated>%v</IsTruncated>", truncated)
	if next != "" {
		fmt.Fprintf(&b, "<NextContinuationToken>%s</NextContinuationToken>", next)
	}
	for i, k := range keys {
		fmt.Fprintf(&b, "<Contents><Key>%s</Key><ETag>&quot;etag-%d&quot;</ETag><Size>10</Size></Contents>", k, i)
	}
	b.WriteString("</ListBucketResult>")
	return b.String()
}

func TestListFindsTheMonthlyParquetFilesOfBothSources(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch {
		case q.Get("list-type") != "2":
			http.Error(w, "not a listing", http.StatusBadRequest)
		case q.Get("prefix") == windPrefix && q.Get("continuation-token") == "":
			_, _ = fmt.Fprint(w, listing(true, "page-2", windPrefix, windPrefix+"RESTRICAO_COFF_EOLICA_2021_10.parquet", windPrefix+"RESTRICAO_COFF_EOLICA_2021_10.csv"))
		case q.Get("prefix") == windPrefix && q.Get("continuation-token") == "page-2":
			_, _ = fmt.Fprint(w, listing(false, "", windPrefix+"RESTRICAO_COFF_EOLICA_2026_10.parquet"))
		case q.Get("prefix") == solarPrefix:
			_, _ = fmt.Fprint(w, listing(false, "", solarPrefix+"DicionarioDados_RestricaoContrainedoff_UsiFotovoltaica.json", solarPrefix+"RESTRICAO_COFF_FOTOVOLTAICA_2024_04.parquet", solarPrefix+"RESTRICAO_COFF_FOTOVOLTAICA_2024_13.parquet"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	files, err := ons.New(srv.URL, srv.Client()).List(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	want := []ingest.File{
		{Key: windPrefix + "RESTRICAO_COFF_EOLICA_2021_10.parquet", Source: curtailment.Wind, Month: time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC), ETag: "etag-1"},
		{Key: windPrefix + "RESTRICAO_COFF_EOLICA_2026_10.parquet", Source: curtailment.Wind, Month: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), ETag: "etag-0"},
		{Key: solarPrefix + "RESTRICAO_COFF_FOTOVOLTAICA_2024_04.parquet", Source: curtailment.Solar, Month: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC), ETag: "etag-1"},
	}
	if !slices.Equal(files, want) {
		t.Errorf("List() =\n%+v\nwant\n%+v", files, want)
	}
}

// column is a record as the ONS stores it, with a column Spillway does not read.
type column struct {
	Subsystem   string            `parquet:"id_subsistema,optional"`
	State       string            `parquet:"id_estado,optional"`
	PlantName   string            `parquet:"nom_usina,optional"`
	Plant       string            `parquet:"id_ons,optional"`
	At          *deprecated.Int96 `parquet:"din_instante,optional"`
	Generated   *float64          `parquet:"val_geracao,optional"`
	Curtailed   *float64          `parquet:"val_geracaonaorealizadaapurada,optional"`
	Reason      *string           `parquet:"cod_razaorestricao,optional"`
	Minutes     *int64            `parquet:"num_minutos_restricao,optional"`
	Description *string           `parquet:"dsc_restricao,optional"`
}

// int96 writes a wall-clock time the way the ONS does.
func int96(t time.Time) *deprecated.Int96 {
	const unixEpochJulianDay = 2_440_588
	day := t.Unix() / 86400
	nanos := uint64(t.Sub(time.Unix(day*86400, 0).UTC()))
	v := deprecated.Int96{uint32(nanos), uint32(nanos >> 32), uint32(day + unixEpochJulianDay)}
	return &v
}

func ptr[T any](v T) *T { return &v }

func parquetFile(t *testing.T, rows []column) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := parquet.NewGenericWriter[column](&buf)
	if _, err := w.Write(rows); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestReadDecodesTheRecords(t *testing.T) {
	file := parquetFile(t, []column{
		{
			Subsystem: "NE", State: "RN", PlantName: "Conj. Assu Sol", Plant: "CJU_RNASSU",
			At:        int96(time.Date(2026, 10, 16, 13, 30, 0, 0, time.UTC)),
			Generated: ptr(120.5), Curtailed: ptr(88.25), Reason: ptr("CNF"), Minutes: ptr(int64(30)),
			Description: ptr("Controle de tensão"),
		},
		// A half hour with no limit: the ONS leaves the curtailment columns empty.
		{Subsystem: "S", State: "RS", PlantName: "Conj. B", Plant: "CJU_B", At: int96(time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)), Generated: ptr(-0.4)},
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+windPrefix+"RESTRICAO_COFF_EOLICA_2026_10.parquet" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(file)
	}))
	t.Cleanup(srv.Close)
	f := ingest.File{Key: windPrefix + "RESTRICAO_COFF_EOLICA_2026_10.parquet", Source: curtailment.Wind}

	var got []curtailment.Record
	err := ons.New(srv.URL, srv.Client()).Read(t.Context(), f, func(r curtailment.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []curtailment.Record{
		{
			Plant: "CJU_RNASSU", PlantName: "Conj. Assu Sol", State: "RN", Subsystem: curtailment.Northeast,
			At:          time.Date(2026, 10, 16, 13, 30, 0, 0, curtailment.Brasilia),
			GeneratedMW: 120.5, CurtailedMW: 88.25, Reason: curtailment.ReasonReliability, LimitedMinutes: 30,
		},
		{Plant: "CJU_B", PlantName: "Conj. B", State: "RS", Subsystem: curtailment.South, At: time.Date(2026, 10, 16, 0, 0, 0, 0, curtailment.Brasilia), GeneratedMW: -0.4},
	}
	if !slices.EqualFunc(got, want, func(a, b curtailment.Record) bool {
		at := a.At.Equal(b.At)
		a.At, b.At = time.Time{}, time.Time{}
		return at && a == b
	}) {
		t.Errorf("Read() =\n%+v\nwant\n%+v", got, want)
	}

	stop := errors.New("stop")
	calls := 0
	err = ons.New(srv.URL, srv.Client()).Read(t.Context(), f, func(curtailment.Record) error {
		calls++
		return stop
	})
	if !errors.Is(err, stop) || calls != 1 {
		t.Errorf("Read() went on past an error: %v after %d records", err, calls)
	}
}

func TestReadRefusesWhatIsNotAFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "missing.parquet") {
			http.NotFound(w, r)
			return
		}
		_, _ = fmt.Fprint(w, "not parquet at all")
	}))
	t.Cleanup(srv.Close)
	c := ons.New(srv.URL, srv.Client())
	ignore := func(curtailment.Record) error { return nil }

	for _, key := range []string{"missing.parquet", "garbage.parquet"} {
		if err := c.Read(t.Context(), ingest.File{Key: key}, ignore); err == nil {
			t.Errorf("Read(%s) = nil", key)
		}
	}
}
