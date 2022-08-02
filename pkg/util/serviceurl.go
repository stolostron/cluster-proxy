package util

import (
	"crypto/sha256"
	"fmt"
)

func GenerateServiceURL(cluster, namespace, service string) string {
	// Using hash to generate a random string
	// URL string should be less than 63 characters
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s_%s_%s", cluster, namespace, service))))[:63]
}
