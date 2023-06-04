PROG_MOUNT_PATH=/sys/fs/bpf

init-bpffs:
	sudo mount -t bpf bpf $(PROG_MOUNT_PATH)
	sudo mkdir -p /sys/fs/bpf/tc/globals

# create map
load-local-dev-map:
	[ -f $(PROG_MOUNT_PATH)/tc/globals/local_dev ] || sudo bpftool map create $(PROG_MOUNT_PATH)/tc/globals/local_dev type hash key 4 value 4 entries 255 name local_dev

load-local-pod-ips-map:
	[ -f $(PROG_MOUNT_PATH)/tc/globals/local_pod_ips ] || sudo bpftool map create $(PROG_MOUNT_PATH)/tc/globals/local_pod_ips type hash key 4 value 24 entries 255 name local_pod_ips

load-cluster-pod-ips-map:
	[ -f $(PROG_MOUNT_PATH)/tc/globals/cluster_pod_ips ] || sudo bpftool map create $(PROG_MOUNT_PATH)/tc/globals/cluster_pod_ips type hash key 4 value 4 entries 255 name cluster_pod_ips

load: init-bpffs load-local-dev-map load-local-pod-ips-map load-cluster-pod-ips-map