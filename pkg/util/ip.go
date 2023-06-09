package util

import (
	"math/big"
	"net"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

func InetIpToUInt32(ip string) uint32 {
	bits := strings.Split(ip, ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])
	var sum uint32
	sum += uint32(b0) << 24
	sum += uint32(b1) << 16
	sum += uint32(b2) << 8
	sum += uint32(b3)
	return sum
}

func InetUint32ToIp(intIP uint32) string {
	var bytes [4]byte
	bytes[0] = byte(intIP & 0xFF)
	bytes[1] = byte((intIP >> 8) & 0xFF)
	bytes[2] = byte((intIP >> 16) & 0xFF)
	bytes[3] = byte((intIP >> 24) & 0xFF)
	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0]).String()
}

func intToIP(i *big.Int) net.IP {
	return net.IP(i.Bytes()).To16()
}

func ipToInt(ip net.IP) *big.Int {
	if v := ip.To4(); v != nil {
		return big.NewInt(0).SetBytes(v)
	}
	return big.NewInt(0).SetBytes(ip.To16())
}

func NextIP(ip net.IP) net.IP {
	i := ipToInt(ip)
	return intToIP(i.Add(i, big.NewInt(1)))
}

func Cmp(ip1, ip2 net.IP) int {
	int1 := ipToInt(ip1)
	int2 := ipToInt(ip2)
	return int1.Cmp(int2)
}

func ParseIPRange(ipRange string) []net.IP {
	arr := strings.Split(ipRange, "-")
	n := len(arr)
	var ips []net.IP
	if n == 1 {
		ips = append(ips, net.ParseIP(arr[0]))
	}

	if n == 2 {
		cur := net.ParseIP(arr[0])
		end := net.ParseIP(arr[1])
		for Cmp(cur, end) <= 0 {
			ips = append(ips, cur)
			cur = NextIP(cur)
		}
	}

	return ips
}

func ExcludeIPs(allIps []net.IP, excludeIps []net.IP) []net.IP {
	resIps := make([]net.IP, 0)
	excludeIpSet := sets.NewString()
	for _, ep := range excludeIps {
		excludeIpSet.Insert(ep.String())
	}
	for _, ap := range allIps {
		if excludeIpSet.Has(ap.String()) {
			continue
		}
		resIps = append(resIps, ap)
	}
	return resIps
}
