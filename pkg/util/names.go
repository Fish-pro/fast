package util

import (
	"fmt"
	"hash/fnv"

	"k8s.io/apimachinery/pkg/util/rand"
)

// GenerateVethName will generate veth name by pod namespace and name
func GenerateVethName(name string, nsAndName string) string {
	hash := fnv.New32a()
	DeepHashObject(hash, nsAndName)
	return fmt.Sprintf("%s-%s", name, rand.SafeEncodeString(fmt.Sprint(hash.Sum32())))
}
