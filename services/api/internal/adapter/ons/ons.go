// Package ons reads the curtailment files the grid operator (ONS) publishes as
// open data: Parquet files in a public S3 bucket, one per source and month,
// which the bucket lists with an ETag that changes whenever a file's content
// does.
package ons

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/deprecated"

	"github.com/christianchiavelli/spillway/services/api/internal/app/ingest"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

// maxFileBytes bounds a download: the largest month so far is under 5 MB.
const maxFileBytes = 64 << 20

// The two datasets, and the names of their monthly Parquet files. The bucket
// holds each month as CSV and XLSX too, and a data dictionary.
var datasets = []struct {
	prefix string
	source curtailment.Source
	file   *regexp.Regexp
}{
	{"dataset/restricao_coff_eolica_tm/", curtailment.Wind, regexp.MustCompile(`/RESTRICAO_COFF_EOLICA_(\d{4})_(\d{2})\.parquet$`)},
	{"dataset/restricao_coff_fotovoltaica_tm/", curtailment.Solar, regexp.MustCompile(`/RESTRICAO_COFF_FOTOVOLTAICA_(\d{4})_(\d{2})\.parquet$`)},
}

// Catalog lists and reads the files. It implements ingest.Catalog.
type Catalog struct {
	base   string
	client *http.Client
}

// New returns a Catalog of the bucket at base.
func New(base string, client *http.Client) *Catalog {
	return &Catalog{base: strings.TrimSuffix(base, "/"), client: client}
}

// listing is the part of S3's ListObjectsV2 answer this package reads.
type listing struct {
	IsTruncated           bool
	NextContinuationToken string
	Contents              []struct {
		Key  string
		ETag string
	}
}

// List returns every monthly file of both sources.
func (c *Catalog) List(ctx context.Context) ([]ingest.File, error) {
	var files []ingest.File
	for _, d := range datasets {
		token := ""
		for {
			q := url.Values{"list-type": {"2"}, "prefix": {d.prefix}}
			if token != "" {
				q.Set("continuation-token", token)
			}
			var page listing
			if err := c.get(ctx, "/?"+q.Encode(), func(body io.Reader) error {
				return xml.NewDecoder(body).Decode(&page)
			}); err != nil {
				return nil, fmt.Errorf("list %s: %w", d.prefix, err)
			}
			for _, o := range page.Contents {
				m := d.file.FindStringSubmatch(o.Key)
				if m == nil {
					continue
				}
				year, _ := strconv.Atoi(m[1])
				month, _ := strconv.Atoi(m[2])
				if month < 1 || month > 12 {
					continue
				}
				files = append(files, ingest.File{
					Key:    o.Key,
					Source: d.source,
					Month:  time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC),
					ETag:   strings.Trim(o.ETag, `"`),
				})
			}
			if !page.IsTruncated || page.NextContinuationToken == "" {
				break
			}
			token = page.NextContinuationToken
		}
	}
	return files, nil
}

// row is a record as the file stores it: the columns Spillway reads, by the
// ONS's names. A missing number reads as zero.
type row struct {
	Subsystem string            `parquet:"id_subsistema,optional"`
	State     string            `parquet:"id_estado,optional"`
	PlantName string            `parquet:"nom_usina,optional"`
	Plant     string            `parquet:"id_ons,optional"`
	At        *deprecated.Int96 `parquet:"din_instante,optional"`
	Generated float64           `parquet:"val_geracao,optional"`
	Curtailed float64           `parquet:"val_geracaonaorealizadaapurada,optional"`
	Reason    string            `parquet:"cod_razaorestricao,optional"`
	Minutes   int64             `parquet:"num_minutos_restricao,optional"`
}

var errNoTime = errors.New("a record with no time")

// Read downloads f whole, since Parquet is read from its end, and passes each
// record to add.
func (c *Catalog) Read(ctx context.Context, f ingest.File, add func(curtailment.Record) error) error {
	var data []byte
	err := c.get(ctx, "/"+f.Key, func(body io.Reader) error {
		var err error
		data, err = io.ReadAll(io.LimitReader(body, maxFileBytes+1))
		if err == nil && len(data) > maxFileBytes {
			err = fmt.Errorf("larger than %d MB", maxFileBytes>>20)
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	file, err := parquet.OpenFile(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open parquet: %w", err)
	}
	r := parquet.NewGenericReader[row](file)
	defer func() { _ = r.Close() }() // it reads bytes in memory: nothing to report
	buf := make([]row, 1024)
	for {
		n, readErr := r.Read(buf)
		for _, x := range buf[:n] {
			if x.At == nil {
				return errNoTime
			}
			err := add(curtailment.Record{
				Plant:          x.Plant,
				PlantName:      x.PlantName,
				State:          x.State,
				Subsystem:      curtailment.Subsystem(x.Subsystem),
				At:             wallClock(*x.At),
				GeneratedMW:    x.Generated,
				CurtailedMW:    x.Curtailed,
				Reason:         curtailment.Reason(x.Reason),
				LimitedMinutes: int(x.Minutes),
			})
			if err != nil {
				return err
			}
		}
		switch {
		case errors.Is(readErr, io.EOF):
			return nil
		case readErr != nil:
			return fmt.Errorf("read parquet: %w", readErr)
		}
	}
}

// wallClock reads an INT96 timestamp, Impala's: the nanoseconds of the day in
// its first eight bytes and the Julian day in the last four. The ONS writes
// Brasília's wall clock in it, as if it were UTC.
func wallClock(v deprecated.Int96) time.Time {
	const unixEpochJulianDay = 2_440_588
	nanos := int64(v[1])<<32 | int64(v[0])
	t := time.Unix((int64(v[2])-unixEpochJulianDay)*24*60*60, nanos).UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), curtailment.Brasilia)
}

func (c *Catalog) get(ctx context.Context, path string, read func(io.Reader) error) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "Spillway (a portfolio project that reads the ONS's open data)")
	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("get: %s", res.Status)
	}
	return read(res.Body)
}
