package driveradminlocal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLocalDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Efs Volume Tools Local Suite")
}
