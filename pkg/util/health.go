package util

import ipamapiv1 "github.com/fast-io/fast/api/proto/v1"

const HealthyOk = "OK"

func IsHealthy(resp *ipamapiv1.HealthResponse) bool {
	return resp.Msg == HealthyOk
}
