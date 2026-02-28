package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		d    Date
		want string
	}{
		{
			name: "normal date",
			d:    Date(time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)),
			want: `"2026-02-28"`,
		},
		{
			name: "time component stripped",
			d:    Date(time.Date(2026, 2, 28, 15, 30, 45, 0, time.UTC)),
			want: `"2026-02-28"`,
		},
		{
			name: "non-UTC timezone normalised",
			d:    Date(time.Date(2026, 6, 15, 0, 0, 0, 0, time.FixedZone("JST", 9*60*60))),
			want: `"2026-06-14"`, // UTC conversion moves it back one day
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.d)
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}
			if got := string(b); got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestDateUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "valid date",
			input: `"2026-02-28"`,
			want:  time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "null is no-op",
			input: `null`,
			want:  time.Time{},
		},
		{
			name:    "wrong format",
			input:   `"28-02-2026"`,
			wantErr: true,
		},
		{
			name:    "not a date string",
			input:   `"hello"`,
			wantErr: true,
		},
		{
			name:    "too short to unquote",
			input:   `"x"`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var d Date
			err := json.Unmarshal([]byte(tc.input), &d)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}
			if tc.input != `null` && !time.Time(d).Equal(tc.want) {
				t.Errorf("got %v, want %v", time.Time(d), tc.want)
			}
		})
	}
}

func TestDateRoundTrip(t *testing.T) {
	original := Date(time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC))
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded Date
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !time.Time(original).Equal(time.Time(decoded)) {
		t.Errorf("round-trip mismatch: %v != %v", time.Time(original), time.Time(decoded))
	}
}

func TestDateValue(t *testing.T) {
	ts := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	d := Date(ts)
	v, err := d.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	got, ok := v.(time.Time)
	if !ok {
		t.Fatalf("Value returned %T, want time.Time", v)
	}
	if !got.Equal(ts) {
		t.Errorf("got %v, want %v", got, ts)
	}
}

func TestDateScan(t *testing.T) {
	t.Run("time.Time source", func(t *testing.T) {
		ts := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
		var d Date
		if err := d.Scan(ts); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if !time.Time(d).Equal(ts) {
			t.Errorf("got %v, want %v", time.Time(d), ts)
		}
	})

	t.Run("unsupported type returns error", func(t *testing.T) {
		var d Date
		if err := d.Scan("2026-02-28"); err == nil {
			t.Fatal("expected error for string source, got nil")
		}
	})

	t.Run("nil source returns error", func(t *testing.T) {
		var d Date
		if err := d.Scan(nil); err == nil {
			t.Fatal("expected error for nil source, got nil")
		}
	})
}
