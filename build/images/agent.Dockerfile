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

FROM ubuntu:20.04 as compiler

WORKDIR /app

ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update &&\
    apt-get install -y git cmake make gcc python3 libncurses-dev gawk flex bison openssl \
    libssl-dev dkms libelf-dev libudev-dev libpci-dev libiberty-dev autoconf

RUN git clone -b v5.4 https://github.com/torvalds/linux.git --depth 1

RUN cd /app/linux/tools/bpf/bpftool && \
    make && make install

FROM ubuntu:20.04

WORKDIR /app

RUN apt-get update && apt-get install -y libelf-dev make sudo clang iproute2 ethtool
COPY --from=compiler /usr/local/sbin/bpftool /usr/local/sbin/bpftool
COPY --from=builder /app/fast-agent /app/fast-agent
COPY bpf bpf
COPY Makefile Makefile

CMD /app/fast-agent
