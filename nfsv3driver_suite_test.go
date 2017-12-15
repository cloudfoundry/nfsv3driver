package nfsv3driver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/nfsv3driver"
	"testing"
	"time"
)

func TestNfsV3Driver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NfsV3Driver Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// override sleep times so that the tests run quickly
	nfsv3driver.PurgeTimeToSleep = time.Microsecond
	return nil
}, func([]byte) {})
