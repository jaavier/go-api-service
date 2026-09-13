package model

import "testing"

func TestPageOffset(t *testing.T) {
	tests := []struct {
		name string
		page Page
		want int
	}{
		{"first page", Page{Number: 1, Size: 20}, 0},
		{"second page", Page{Number: 2, Size: 20}, 20},
		{"third page small size", Page{Number: 3, Size: 5}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.page.Offset(); got != tt.want {
				t.Errorf("Offset() = %d, want %d", got, tt.want)
			}
		})
	}
}
