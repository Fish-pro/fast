package nettools

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/vishvananda/netlink"
)

func RandomVethName() (string, error) {
	entropy := make([]byte, 4)
	_, err := rand.Read(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate random veth name: %w", err)
	}

	// NetworkManager (recent versions) will ignore veth devices that start with "veth"
	return fmt.Sprintf("veth%x", entropy), nil
}

func CreateVethPair(ifname string, mtu int, hostname ...string) (*netlink.Veth, *netlink.Veth, error) {
	var pairName string
	if len(hostname) != 0 {
		pairName = hostname[0]
	} else {
		for {
			randName, err := RandomVethName()
			pairName = randName
			if err != nil {
				return nil, nil, err
			}
			if _, err := netlink.LinkByName(pairName); err != nil && !os.IsExist(err) {
				break
			}
		}
	}
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: ifname,
			MTU:  mtu,
		},
		PeerName: pairName,
	}
	if err := netlink.LinkAdd(veth); err != nil {
		return nil, nil, err
	}

	veth1, err := netlink.LinkByName(ifname)
	if err != nil {
		netlink.LinkDel(veth)
		return nil, nil, err
	}

	veth2, err := netlink.LinkByName(pairName)
	if err != nil {
		netlink.LinkDel(veth)
		return nil, nil, err
	}

	return veth1.(*netlink.Veth), veth2.(*netlink.Veth), nil
}

func CreateArpEntry(ip, mac, dev string) error {
	processInfo := exec.Command(
		"/bin/sh", "-c",
		fmt.Sprintf("arp -s %s %s -i %s", ip, mac, dev),
	)
	_, err := processInfo.Output()
	return err
}

func DeleteArpEntry(ip, dev string) error {
	processInfo := exec.Command(
		"/bin/sh", "-c",
		fmt.Sprintf("arp -d %s -i %s", ip, dev),
	)
	_, err := processInfo.Output()
	return err
}

func CreateVxlanAndUp(name string) (*netlink.Vxlan, error) {
	l, _ := netlink.LinkByName(name)

	vxlan, ok := l.(*netlink.Vxlan)
	if ok && vxlan != nil {
		return vxlan, nil
	}

	processInfo := exec.Command(
		"/bin/sh", "-c",
		fmt.Sprintf("ip link add name %s type vxlan external", name),
	)
	if _, err := processInfo.Output(); err != nil {
		return nil, err
	}

	l, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}

	vxlan, ok = l.(*netlink.Vxlan)
	if !ok {
		return nil, fmt.Errorf("found the device %q but it's not a vxlan", name)
	}
	if err = netlink.LinkSetUp(vxlan); err != nil {
		return nil, fmt.Errorf("set up vxlan %q error, err: %w", name, err)
	}
	return vxlan, nil
}

func DeviceExistIp(link netlink.Link) (string, error) {
	dev, err := net.InterfaceByIndex(link.Attrs().Index)
	if err == nil {
		addrs, err := dev.Addrs()
		if err == nil && len(addrs) > 0 {
			str := addrs[0].String()
			tmpIp := strings.Split(str, "/")
			if len(tmpIp) == 2 && net.ParseIP(tmpIp[0]).To4() != nil {
				return str, nil
			}
		}
	}
	return "", nil
}
