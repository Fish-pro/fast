package util

import "testing"

func TestGenerateVethName(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name      string
		prefix    string
		nsAndName string
		want      string
	}{
		{
			name:      "case 1",
			prefix:    "fast",
			nsAndName: "default/test",
			want:      "fast-f67f74bcf",
		},
		{
			name:      "case 2",
			prefix:    "fast",
			nsAndName: "default/test",
			want:      "fast-f67f74bcf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateVethName(tt.prefix, tt.nsAndName); got != tt.want {
				t.Errorf("GenerateVethName() = %v, want %v", got, tt.want)
			}
		})
	}
}
