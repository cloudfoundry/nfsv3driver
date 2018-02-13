package nfsv3driver_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"

	"code.cloudfoundry.org/goshims/ldapshim/ldap_fake"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/nfsv3driver"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"gopkg.in/ldap.v2"
)

var _ = Describe("IdResolverTest", func() {
	var ldapFake *ldap_fake.FakeLdap
	var ldapConnectionFake *ldap_fake.FakeLdapConnection
	var ldapIdResolver nfsv3driver.IdResolver
	var env voldriver.Env
	var uid string
	var gid string
	var err error
	var ldapTimeout time.Duration

	Context("when the connection is successful", func() {

		BeforeEach(func() {
			ldapFake = &ldap_fake.FakeLdap{}
			ldapConnectionFake = &ldap_fake.FakeLdapConnection{}
			ldapFake.DialReturns(ldapConnectionFake, nil)
			logger := lagertest.NewTestLogger("nfs-mounter")
			testContext := context.TODO()
			env = driverhttp.NewHttpDriverEnv(logger, testContext)
			ldapTimeout = 120 * time.Second

			ldapIdResolver = nfsv3driver.NewLdapIdResolver("svcuser", "svcpw", "host", 111, "tcp", "cn=Users,dc=test,dc=com", ldapFake, ldapTimeout)
		})

		JustBeforeEach(func() {
			uid, gid, err = ldapIdResolver.Resolve(env, "user", "pw")
		})

		Context("when search returns sucessfully", func() {
			BeforeEach(func() {
				entry := &ldap.Entry{
					DN: "foo",
					Attributes: []*ldap.EntryAttribute{
						&ldap.EntryAttribute{Name: "uidNumber", Values: []string{"100"}},
						&ldap.EntryAttribute{Name: "gidNumber", Values: []string{"100"}},
					},
				}

				result := &ldap.SearchResult{
					Entries: []*ldap.Entry{entry},
				}

				ldapConnectionFake.SearchReturns(result, nil)
			})

			It("set timeout for connection", func() {
				Expect(ldapConnectionFake.SetTimeoutCallCount()).To(Equal(1))
			})

			It("does not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the expected UID and GID", func() {
				Expect(uid).To(Equal("100"))
				Expect(gid).To(Equal("100"))
			})

			Context("when the credentials are not good", func() {
				BeforeEach(func() {
					ldapConnectionFake.BindStub = func(u, p string) error {
						if u == "svcuser" {
							return nil
						} else {
							return errors.New("badness")
						}
					}
				})
				It("should find the user and then fail", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("badness"))
					Expect(ldapConnectionFake.SearchCallCount()).To(Equal(1))
					Expect(uid).To(BeEmpty())
				})
			})
		})

		Context("when the search returns empty", func() {
			BeforeEach(func() {
				result := &ldap.SearchResult{Entries: []*ldap.Entry{}}
				ldapConnectionFake.SearchReturns(result, nil)
			})

			It("reports an error for the missing user", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("User does not exist"))
			})
		})

		Context("when the search returns multiple results", func() {
			BeforeEach(func() {
				entry := &ldap.Entry{
					DN: "foo",
					Attributes: []*ldap.EntryAttribute{
						&ldap.EntryAttribute{Name: "uidNumber", Values: []string{"100"}},
						&ldap.EntryAttribute{Name: "gidNumber", Values: []string{"100"}},
					},
				}

				result := &ldap.SearchResult{
					Entries: []*ldap.Entry{entry, entry},
				}

				ldapConnectionFake.SearchReturns(result, nil)
			})

			It("reports an error for the ambiguous search", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Ambiguous search"))
			})
		})
	})
})
