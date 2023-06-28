FROM --platform=linux/amd64 golang:1.20 as builder

WORKDIR /app
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

COPY cmd cmd
COPY pkg pkg
COPY api api
COPY bpf bpf
COPY vendor vendor

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o fast-agent cmd/agent/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o fastctl cmd/fastctl/main.go

FROM ubuntu:20.04 as compiler

WORKDIR /app

ARG DEBIAN_FRONTEND=noninteractive

COPY bpf/* .

RUN apt-get update && apt-get install -y git cmake make gcc python3 libncurses-dev gawk flex bison openssl \
    libssl-dev dkms libelf-dev libudev-dev libpci-dev libiberty-dev autoconf \
    clang llvm libelf-dev libbpf-dev bpfcc-tools libbpfcc-dev

RUN git clone -b v5.4 https://github.com/torvalds/linux.git --depth 1
RUN cd /app/linux/tools/bpf/bpftool && \
    make && make install

RUN clang -g  -O2 -emit-llvm -c vxlan_egress.c -o - | llc -march=bpf -filetype=obj -o vxlan_egress.o
RUN clang -g  -O2 -emit-llvm -c vxlan_ingress.c -o - | llc -march=bpf -filetype=obj -o vxlan_ingress.o
RUN clang -g  -O2 -emit-llvm -c veth_ingress.c -o - | llc -march=bpf -filetype=obj -o veth_ingress.o

FROM ubuntu:20.04

WORKDIR /app

RUN apt-get update && apt-get install -y libelf-dev make sudo clang iproute2 ethtool
COPY bpf bpf
COPY Makefile Makefile
COPY --from=compiler /usr/local/sbin/bpftool /usr/local/sbin/bpftool
COPY --from=compiler /app/vxlan_egress.o /app/bpf/vxlan_egress.o
COPY --from=compiler /app/vxlan_ingress.o /app/bpf/vxlan_ingress.o
COPY --from=compiler /app/veth_ingress.o /app/bpf/veth_ingress.o
COPY --from=builder /app/fastctl /usr/local/bin/fastctl
COPY --from=builder /app/fast-agent /app/fast-agent

CMD /app/fast-agent
