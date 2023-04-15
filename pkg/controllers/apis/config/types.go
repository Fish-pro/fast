package config

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmconfig "k8s.io/controller-manager/config"
)

// FastControllerManagerConfiguration contains elements describing firefly-controller manager.
type FastControllerManagerConfiguration struct {
	metav1.TypeMeta

	// Generic holds configuration for a generic controller-manager
	Generic cmconfig.GenericControllerManagerConfiguration
}
