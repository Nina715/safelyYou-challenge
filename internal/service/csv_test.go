package service

import (
	"strings"
	"testing"
)

func TestParseCSV_Valid(t *testing.T) {
	input := "device_id,name\ndevice-1,Alpha\ndevice-2,Beta\n"
	ids, err := parseCSV(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("got %d ids, want 2", len(ids))
	}
	if ids[0] != "device-1" || ids[1] != "device-2" {
		t.Errorf("ids: got %v", ids)
	}
}

func TestParseCSV_MissingColumn(t *testing.T) {
	input := "name,model\nAlpha,X1\n"
	_, err := parseCSV(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for missing device_id column")
	}
}

func TestParseCSV_Deduplication(t *testing.T) {
	input := "device_id\ndev-1\ndev-1\ndev-2\n"
	ids, err := parseCSV(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("got %d ids, want 2 (duplicates not removed)", len(ids))
	}
}

func TestParseCSV_SkipsEmptyRows(t *testing.T) {
	input := "device_id\ndev-1\n\ndev-2\n"
	ids, err := parseCSV(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("got %d ids, want 2", len(ids))
	}
}

func TestParseCSV_EmptyFile(t *testing.T) {
	_, err := parseCSV(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestParseCSV_HeaderOnly(t *testing.T) {
	input := "device_id\n"
	ids, err := parseCSV(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("got %d ids, want 0", len(ids))
	}
}
