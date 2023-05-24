package bpf_map

type LocalDevType uint32

const (
	VxlanDevType LocalDevType = 1
	VethDevType  LocalDevType = 2
)

type LocalDevMapKey struct {
	Type LocalDevType
}

type LocalDevMapValue struct {
	IfIndex uint32
}

type LocalIpsMapKey struct {
	IP uint32
}

type LocalIpsMapInfo struct {
	IfIndex    uint32
	LxcIfIndex uint32

	MAC     [8]byte
	NodeMAC [8]byte
}

type ClusterIpsMapKey struct {
	IP uint32
}

type ClusterIpsMapInfo struct {
	IP uint32
}
