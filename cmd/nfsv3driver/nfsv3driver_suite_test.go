package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"testing"
	"time"
)

func TestNfsV3Driver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NFS V3 Main Suite")
}

var driverPath string

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Second)

	// this test suite shares an os environment and therefore cannot run in parallel
	configuration, _ := GinkgoConfiguration()
	Expect(configuration.ParallelTotal).To(Equal(1))

	var err error
	driverPath, err = Build("code.cloudfoundry.org/nfsv3driver/cmd/nfsv3driver")
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	CleanupBuildArtifacts()
})
