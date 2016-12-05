package nfsv3driver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestNfsV3Driver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NfsV3Driver Suite")
}
