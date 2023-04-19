package util

import v1 "k8s.io/api/core/v1"

func IsPodAlive(pod *v1.Pod) bool {
	if pod == nil {
		return false
	}

	if pod.DeletionTimestamp != nil {
		return false
	}

	if pod.Status.Phase == v1.PodSucceeded && pod.Spec.RestartPolicy != v1.RestartPolicyAlways {
		return false
	}

	if pod.Status.Phase == v1.PodFailed && pod.Spec.RestartPolicy == v1.RestartPolicyNever {
		return false
	}

	if pod.Status.Phase == v1.PodFailed && pod.Status.Reason == "Evicted" {
		return false
	}

	return true
}
