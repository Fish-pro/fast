package bpf_map

import "testing"

func TestPrintMapSize(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PrintMapSize()
		})
	}
}
