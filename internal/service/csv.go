package service

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func LoadFromCSV(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	return parseCSV(f)
}

func parseCSV(r io.Reader) ([]string, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	deviceIDCol := -1
	for i, col := range header {
		if strings.EqualFold(strings.TrimSpace(col), "device_id") {
			deviceIDCol = i
			break
		}
	}
	if deviceIDCol == -1 {
		return nil, errors.New("csv missing required column: device_id")
	}

	var ids []string
	seen := make(map[string]struct{})
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		if deviceIDCol >= len(row) {
			continue
		}
		id := strings.TrimSpace(row[deviceIDCol])
		if id == "" {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}
