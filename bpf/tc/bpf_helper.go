package tc

import (
	"fmt"
)

type BpfTcDirectType string

const (
	ProgramDefaultPath                 = "/opt/fast"
	IngressType        BpfTcDirectType = "ingress"
	EgressType         BpfTcDirectType = "egress"
)

func GetVethIngressPath() string {
	return ProgramDefaultPath + "/veth_ingress.o"
}

func GetVxlanIngressPath() string {
	return ProgramDefaultPath + "/vxlan_ingress.o"
}

func GetVxlanEgressPath() string {
	return ProgramDefaultPath + "/vxlan_egress.o"
}

func TryAttachBPF(dev string, direct BpfTcDirectType, program string) error {
	if !ExistClsact(dev) {
		err := AddClsactQdiscIntoDev(dev)
		if err != nil {
			return err
		}
	}

	switch direct {
	case IngressType:
		if ExistIngress(dev) {
			return nil
		}
		return AttachIngressBPFIntoDev(dev, program)
	case EgressType:
		if ExistEgress(dev) {
			return nil
		}
		return AttachEgressBPFIntoDev(dev, program)
	}
	return fmt.Errorf("unknow error occurred in TryAttachBPF")
}

func DetachBPF(dev string) error {
	return DelClsactQdiscIntoDev(dev)
}
