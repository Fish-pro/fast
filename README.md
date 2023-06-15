# Fast

Fast is a Kubernetes CNI based on  [eBPF](https://ebpf.io)  implementation

## Architecture

![fast](images/fast.png)

Components:
+ fast-cni
  + implement CNI capabilities
  + access fast-agent fetch pod IP and store to the node local eBPF map
  + attach eBPF program to NIC
+ fast-agent
  + the interface that implements IP allocation
  + obtain the cluster pod IP and store the information to the cluster eBPF map
  + init eBPF map
+ fast-controller-manager
  + custom resources control
  + gc management to prevent IP leakage

## Quick Start

//TODO

## What's Next

More will be coming Soon. Welcome to [open an issue](https://github.com/Fish-pro/fast/issues) and [propose a PR](https://github.com/Fish-pro/fast/pulls). 🎉🎉🎉

## Contributors

<a href="https://github.com/Fish-pro/fast/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=Fish-pro/fast" />
</a>

Made with [contrib.rocks](https://contrib.rocks).

## License

Fast is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.