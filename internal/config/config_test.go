package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		env         map[string]string
		wantPort    int
		wantWorkers int
		wantDebug   bool
		wantErr     bool
	}{
		{
			name:        "defaults",
			env:         map[string]string{"DATABASE_URL": "postgres://localhost/test"},
			wantPort:    8080,
			wantWorkers: 4,
			wantDebug:   false,
		},
		{
			name: "custom port and workers",
			env: map[string]string{
				"DATABASE_URL": "postgres://localhost/test",
				"PORT":         "9090",
				"MAX_WORKERS":  "8",
				"DEBUG":        "true",
			},
			wantPort:    9090,
			wantWorkers: 8,
			wantDebug:   true,
		},
		{
			name:    "missing DATABASE_URL",
			env:     map[string]string{},
			wantErr: true,
		},
		{
			name: "invalid PORT",
			env: map[string]string{
				"DATABASE_URL": "postgres://localhost/test",
				"PORT":         "not-a-number",
			},
			wantErr: true,
		},
		{
			name: "invalid MAX_WORKERS",
			env: map[string]string{
				"DATABASE_URL": "postgres://localhost/test",
				"MAX_WORKERS":  "abc",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Clear vars that might bleed across sub-tests.
			for _, k := range []string{"DATABASE_URL", "PORT", "MAX_WORKERS", "DEBUG"} {
				t.Setenv(k, "")
			}
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			cfg, err := Load()
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Port != tc.wantPort {
				t.Errorf("Port: got %d, want %d", cfg.Port, tc.wantPort)
			}
			if cfg.MaxWorkers != tc.wantWorkers {
				t.Errorf("MaxWorkers: got %d, want %d", cfg.MaxWorkers, tc.wantWorkers)
			}
			if cfg.Debug != tc.wantDebug {
				t.Errorf("Debug: got %v, want %v", cfg.Debug, tc.wantDebug)
			}
		})
	}
}
