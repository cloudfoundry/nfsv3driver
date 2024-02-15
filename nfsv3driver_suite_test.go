package nfsv3driver_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/nfsv3driver"
	"testing"
	"time"
)

func TestNfsV3Driver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NFS V3 Driver Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	return nil
}, func([]byte) {
	nfsv3driver.PurgeTimeToSleep = time.Microsecond
})
