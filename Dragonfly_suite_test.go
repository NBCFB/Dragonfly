package Dragonfly_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDragonfly(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dragonfly Suite")
}
