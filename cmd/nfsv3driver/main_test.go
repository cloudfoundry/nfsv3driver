package main_test

import (
	"io/ioutil"
	"net"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os"
)

var _ = Describe("Main", func() {
	var (
		session *gexec.Session
		command *exec.Cmd
		err     error
	)

	BeforeEach(func() {
		command = exec.Command(driverPath)
	})

	JustBeforeEach(func() {
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		session.Kill().Wait()
	})

	Context("with a driver path", func() {
		BeforeEach(func() {
			dir, err := ioutil.TempDir("", "driversPath")
			Expect(err).ToNot(HaveOccurred())

			command.Args = append(command.Args, "-driversPath="+dir)
		})

		It("listens on tcp/7589 by default", func() {
			EventuallyWithOffset(1, func() error {
				_, err := net.Dial("tcp", "0.0.0.0:7589")
				return err
			}, 5).ShouldNot(HaveOccurred())
		})

		Context("in another context", func() {
			BeforeEach(func() {
				command.Args = append(command.Args, "-listenAddr=0.0.0.0:7591")
				command.Args = append(command.Args, "-adminAddr=0.0.0.0:7592")
			})

			It("listens on tcp/7590 for admin reqs by default", func() {

				EventuallyWithOffset(1, func() error {
					_, err := net.Dial("tcp", "0.0.0.0:7592")
					return err
				}, 5).ShouldNot(HaveOccurred())
			})
		})

		Context("given correct LDAP arguments set in the environment", func() {
			BeforeEach(func() {
				os.Setenv("LDAP_SVC_USER", "user")
				os.Setenv("LDAP_SVC_PASS", "password")
				os.Setenv("LDAP_USER_FQDN", "cn=Users,dc=corp,dc=testdomain,dc=com")
				os.Setenv("LDAP_HOST", "ldap.testdomain.com")
				os.Setenv("LDAP_PORT", "389")
				os.Setenv("LDAP_PROTO", "tcp")
				command.Args = append(command.Args, "-listenAddr=0.0.0.0:7593")
				command.Args = append(command.Args, "-adminAddr=0.0.0.0:7594")
			})
			It("listens on tcp/7589 by default", func() {
				EventuallyWithOffset(1, func() error {
					_, err := net.Dial("tcp", "0.0.0.0:7593")
					return err
				}, 5).ShouldNot(HaveOccurred())
			})
		})
		Context("given incomplete LDAP arguments set in the environment", func() {
			BeforeEach(func() {
				os.Setenv("LDAP_HOST", "ldap.testdomain.com")
				os.Setenv("LDAP_PORT", "389")
				os.Setenv("LDAP_PROTO", "tcp")
				command.Args = append(command.Args, "-listenAddr=0.0.0.0:7595")
				command.Args = append(command.Args, "-adminAddr=0.0.0.0:7596")
			})
			It("fails to start", func() {
				EventuallyWithOffset(1, func() error {
					_, err := net.Dial("tcp", "0.0.0.0:7595")
					return err
				}, 5).Should(HaveOccurred())
			})
		})

	})
})
