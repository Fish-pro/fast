package util

import "testing"

func TestInetIpToUInt32(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want uint32
	}{
		{
			name: "case 1",
			ip:   "10.244.10.23",
			want: uint32(183765527),
		},
		{
			name: "case 2",
			ip:   "10.29.15.49",
			want: uint32(169676593),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InetIpToUInt32(tt.ip); got != tt.want {
				t.Errorf("InetIpToUInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInetUint32ToIp(t *testing.T) {
	tests := []struct {
		name  string
		intIP uint32
		want  string
	}{
		{
			name:  "case 1",
			intIP: uint32(183765527),
			want:  "10.244.10.23",
		},
		{
			name:  "case 2",
			intIP: uint32(169676593),
			want:  "10.29.15.49",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InetUint32ToIp(tt.intIP); got != tt.want {
				t.Errorf("InetUint32ToIp() = %v, want %v", got, tt.want)
			}
		})
	}
}
