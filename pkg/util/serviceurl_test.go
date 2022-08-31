package util

import (
	"fmt"
	"testing"
)

func TestGenerateServiceURL(t *testing.T) {
	fmt.Println(GenerateServiceURL("local-cluster", "default", "local-hello"))
}
