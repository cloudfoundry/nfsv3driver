package nfsv3driver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/nfsv3driver"
	"code.cloudfoundry.org/goshims/ldapshim/ldap_fake"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/lager/lagertest"
	"context"
	"gopkg.in/ldap.v2"
)

var _ = FDescribe("IdResolverTest", func() {
  var ldapFake *ldap_fake.FakeLdap
	var ldapConnectionFake *ldap_fake.FakeLdapConnection
	var ldapIdResolver nfsv3driver.IdResolver
	var env voldriver.Env
	var uid string
	var gid string
	var err error

	Context("when the connection is successful", func() {

		BeforeEach(func() {
			ldapFake = &ldap_fake.FakeLdap{}
			ldapConnectionFake = &ldap_fake.FakeLdapConnection{}
			ldapFake.DialReturns(ldapConnectionFake, nil)
			logger := lagertest.NewTestLogger("nfs-mounter")
			testContext := context.TODO()
			env = driverhttp.NewHttpDriverEnv(logger, testContext)

			ldapIdResolver = nfsv3driver.NewLdapIdResolver("user", "pw", "host", 111, "tcp", "cn=Users,dc=test,dc=com", ldapFake)
		})

		JustBeforeEach(func(){
			uid, gid, err = ldapIdResolver.Resolve(env, "user", "pw")
		})

		Context("when search returns sucessfully", func(){
			BeforeEach(func(){
				entry := &ldap.Entry{
					DN: "foo",
					Attributes: []*ldap.EntryAttribute{},
				}

				result := &ldap.SearchResult{
					Entries: []*ldap.Entry{entry},
				}

				ldapConnectionFake.SearchReturns(result, nil)
			})

			It("does not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

		})

	})
})
