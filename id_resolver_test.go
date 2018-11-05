package nfsv3driver_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/goshims/ldapshim/ldap_fake"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/nfsv3driver"
	"gopkg.in/ldap.v2"
)

var _ = Describe("IdResolverTest", func() {
	var ldapFake *ldap_fake.FakeLdap
	var ldapConnectionFake *ldap_fake.FakeLdapConnection
	var ldapIdResolver nfsv3driver.IdResolver
	var env dockerdriver.Env
	var uid string
	var gid string
	var err error
	var ldapCACert string
	var ldapTimeout time.Duration

	Context("when the connection is successful", func() {
		BeforeEach(func() {
			ldapFake = &ldap_fake.FakeLdap{}
			ldapConnectionFake = &ldap_fake.FakeLdapConnection{}
			ldapFake.DialReturns(ldapConnectionFake, nil)
			logger := lagertest.NewTestLogger("nfs-mounter")
			testContext := context.TODO()
			env = driverhttp.NewHttpDriverEnv(logger, testContext)
			ldapCACert = ""
			ldapTimeout = 120 * time.Second
			ldapConnectionFake.SearchReturns(&ldap.SearchResult{}, nil)
		})

		JustBeforeEach(func() {
			ldapIdResolver = nfsv3driver.NewLdapIdResolver(
				"svcuser",
				"svcpw",
				"host",
				111,
				"tcp",
				"cn=Users,dc=test,dc=com",
				ldapCACert,
				ldapFake,
				ldapTimeout,
			)
			uid, gid, err = ldapIdResolver.Resolve(env, "user", "pw")
		})

		Context("when CA cert is provided", func() {
			BeforeEach(func() {
				ldapCACert = `-----BEGIN CERTIFICATE-----
MIIDGTCCAgGgAwIBAgIRAIlVvSGFPY1EvNayuTpPAScwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAeFw0xODA1MzExNzU5MTBaFw0xOTA1MzExNzU5
MTBaMBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQCSf8J68FYrRuE8+NumcleeI10+O5QGibQ3+axX79eFS3RGcQKn5UOr
OFE/RM/ghc7sUD8urLhlA2QAua+0dZEr+QtNswDxLfWljw08azR4xkPnBejdwYKU
jHHU9UoJrxEgWqNFwTWWCyHYERUK/RFSrSUJaZLv1fRa9C+wbkD2Wd+aesPU6TZr
5f6DT1UdL5umykwVoKy9ymA1CUi3iRSPuIxF0iuwwNtgtS0Dswi9+gqICOYp+lGJ
RM2zRZFas8clubvkIRYlO2YG8hb181uxW9nLAfUfJjjtDt7lp5z/eZqliFwzrl0i
DG8xWUppHV9654hGRDOL2ow3u8kwNv9/AgMBAAGjajBoMA4GA1UdDwEB/wQEAwIC
pDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MDAGA1UdEQQp
MCeCJW5mc3Rlc3RsZGFwc2VydmVyLnNlcnZpY2UuY2YuaW50ZXJuYWwwDQYJKoZI
hvcNAQELBQADggEBAB/4KT3+G5YqrnCCF+GmYlxZO9ScRA6yPBtwXTQe7WH8Yfz2
bnUs4jKhK2wh3+RSTsBwV9afF+xm/uVrD9iZveixC1E3NqJwlchHc2bv9NCvC8OY
VShIx+8Joqpud6VIrzclhus2lo9Dvn55at3Z/5SYDf07fDmSJ5pZuLUVryiJk9AT
G0GELNbBftMakAJaH6eqGvcNbDRMeFqq7VyjthQJRPWSaWKA6TsfzgiO9lwx1wd1
1ZtN1nl1NexFqcan26vg0f1SwLM9r9mVXrKII/T60RXKvtcAkMS3XfaebG3ulout
z6sbK6WkL0AwPEcI/HzUOrsAUBtyY8cfy6yVcuQ=
-----END CERTIFICATE-----`
				ldapFake.DialTLSReturns(ldapConnectionFake, nil)
			})

			It("connects via TLS", func() {
				Expect(ldapFake.DialTLSCallCount()).To(Equal(1))
				protocol, addr, config := ldapFake.DialTLSArgsForCall(0)
				Expect(protocol).To(Equal("tcp"))
				Expect(addr).To(Equal("host:111"))
				Expect(config.ServerName).To(Equal("host"))
				Expect(config.RootCAs.Subjects()).To(HaveLen(1))
				Expect(string(config.RootCAs.Subjects()[0])).To(ContainSubstring("Acme Co"))
			})
		})

		Context("when CA cert is not provided", func() {
			BeforeEach(func() {
				ldapCACert = ""
			})

			It("connects without TLS", func() {
				Expect(ldapFake.DialCallCount()).To(Equal(1))
				protocol, addr := ldapFake.DialArgsForCall(0)
				Expect(protocol).To(Equal("tcp"))
				Expect(addr).To(Equal("host:111"))
			})
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

		Context("when search does not return GID", func() {
			BeforeEach(func() {
				entry := &ldap.Entry{
					DN: "foo",
					Attributes: []*ldap.EntryAttribute{
						&ldap.EntryAttribute{Name: "uidNumber", Values: []string{"100"}},
					},
				}

				result := &ldap.SearchResult{
					Entries: []*ldap.Entry{entry},
				}

				ldapConnectionFake.SearchReturns(result, nil)
			})

			It("does not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("sets GID same as UID", func() {
				Expect(uid).To(Equal("100"))
				Expect(gid).To(Equal("100"))
			})
		})

		Context("when search returns empty GID", func() {
			BeforeEach(func() {
				entry := &ldap.Entry{
					DN: "foo",
					Attributes: []*ldap.EntryAttribute{
						&ldap.EntryAttribute{Name: "uidNumber", Values: []string{"100"}},
						&ldap.EntryAttribute{Name: "gidNumber", Values: []string{""}},
					},
				}

				result := &ldap.SearchResult{
					Entries: []*ldap.Entry{entry},
				}

				ldapConnectionFake.SearchReturns(result, nil)
			})

			It("does not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("sets GID same as UID", func() {
				Expect(uid).To(Equal("100"))
				Expect(gid).To(Equal("100"))
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
