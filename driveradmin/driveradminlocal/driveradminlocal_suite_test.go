package driveradminlocal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLocalDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Driver Admin API Local Suite")
}
