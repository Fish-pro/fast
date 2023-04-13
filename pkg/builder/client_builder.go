/*
Copyright 2023 The fast Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package builder

import (
	restclient "k8s.io/client-go/rest"
	"k8s.io/controller-manager/pkg/clientbuilder"
	"k8s.io/klog/v2"

	ipsversioned "github.com/fast-io/fast/pkg/generated/clientset/versioned"
)

// IpsControllerClientBuilder allows you to get clients and configs for application controllers
type IpsControllerClientBuilder interface {
	clientbuilder.ControllerClientBuilder
	IpsClient(name string) (ipsversioned.Interface, error)
	IpsClientOrDie(name string) ipsversioned.Interface
}

// make sure that SimpleIpsControllerClientBuilder implements IpsControllerClientBuilder
var _ IpsControllerClientBuilder = SimpleIpsControllerClientBuilder{}

// NewSimpleIpsControllerClientBuilder creates a SimpleIpsControllerClientBuilder
func NewSimpleIpsControllerClientBuilder(config *restclient.Config) SimpleIpsControllerClientBuilder {
	return SimpleIpsControllerClientBuilder{
		clientbuilder.SimpleControllerClientBuilder{
			ClientConfig: config,
		},
	}
}

// SimpleIpsControllerClientBuilder returns a fixed client with different user agents
type SimpleIpsControllerClientBuilder struct {
	clientbuilder.SimpleControllerClientBuilder
}

// IpsClient returns a versioned.Interface built from the ClientBuilder
func (b SimpleIpsControllerClientBuilder) IpsClient(name string) (ipsversioned.Interface, error) {
	clientConfig, err := b.Config(name)
	if err != nil {
		return nil, err
	}
	return ipsversioned.NewForConfig(clientConfig)
}

// IpsClientOrDie returns a versioned.interface built from the ClientBuilder with no error.
// If it gets an error getting the client, it will log the error and kill the process it's running in.
func (b SimpleIpsControllerClientBuilder) IpsClientOrDie(name string) ipsversioned.Interface {
	client, err := b.IpsClient(name)
	if err != nil {
		klog.Fatal(err)
	}
	return client
}
