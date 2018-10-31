package nfsv3driver_test

import (
	"context"
	"fmt"
	"strings"

	"code.cloudfoundry.org/goshims/ioutilshim/ioutil_fake"
	"code.cloudfoundry.org/goshims/osshim/os_fake"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/nfsv3driver"
	"code.cloudfoundry.org/nfsv3driver/nfsdriverfakes"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/voldriver/voldriverfakes"
	"code.cloudfoundry.org/volumedriver"
	nfsfakes "code.cloudfoundry.org/volumedriver/volumedriverfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NfsV3Mounter", func() {

	var (
		logger      lager.Logger
		testContext context.Context
		env         voldriver.Env
		err         error

		fakeInvoker      *voldriverfakes.FakeInvoker
		fakeIdResolver   *nfsdriverfakes.FakeIdResolver
		fakeIoutil       *ioutil_fake.FakeIoutil
		fakeOs           *os_fake.FakeOs
		fakeMountChecker *nfsfakes.FakeMountChecker

		subject volumedriver.Mounter

		opts map[string]interface{}
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("nfs-mounter")
		testContext = context.TODO()
		env = driverhttp.NewHttpDriverEnv(logger, testContext)
		opts = map[string]interface{}{}

		fakeInvoker = &voldriverfakes.FakeInvoker{}
		fakeIoutil = &ioutil_fake.FakeIoutil{}
		fakeOs = &os_fake.FakeOs{}
		fakeMountChecker = &nfsfakes.FakeMountChecker{}
		fakeMountChecker.ExistsReturns(true, nil)

		source := nfsv3driver.NewNfsV3ConfigDetails()
		source.ReadConf("uid,gid", "", []string{})

		mounts := nfsv3driver.NewNfsV3ConfigDetails()
		mounts.ReadConf("sloppy_mount,allow_other,allow_root,multithread,default_permissions,fusenfs_uid,fusenfs_gid,username,password", "", []string{})

		subject = nfsv3driver.NewNfsV3Mounter(fakeInvoker, fakeOs, fakeIoutil, fakeMountChecker, nfsv3driver.NewNfsV3Config(source, mounts), nil)
	})

	Context("#Mount", func() {
		Context("when mount succeeds", func() {
			JustBeforeEach(func() {
				fakeInvoker.InvokeReturns(nil, nil)
				err = subject.Mount(env, "source", "target", opts)
			})

			It("should return without error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should use the passed in variables", func() {
				_, cmd, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("fuse-nfs"))
				Expect(strings.Join(args, " ")).To(ContainSubstring("-n source"))
				Expect(strings.Join(args, " ")).To(ContainSubstring("-m target"))
			})

			Context("when mounting read only", func() {
				BeforeEach(func() {
					opts["readonly"] = true
				})

				It("should not include the -O flag, since garden handles readonly mounts", func() {
					_, _, args := fakeInvoker.InvokeArgsForCall(0)
					Expect(strings.Join(args, " ")).NotTo(ContainSubstring("-O"))
				})
			})
		})

		Context("when mount errors", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("error"))

				err = subject.Mount(env, "source", "target", opts)
			})

			It("should return without error", func() {
				Expect(err).To(HaveOccurred())
				_, ok := err.(voldriver.SafeError)
				Expect(ok).To(BeTrue())
			})
		})

		Context("when mount is cancelled", func() {
			// TODO: when we pick up the lager.Context
		})

		Context("when username mapping is enabled", func() {
			BeforeEach(func() {
				fakeIdResolver = &nfsdriverfakes.FakeIdResolver{}

				source := nfsv3driver.NewNfsV3ConfigDetails()
				source.ReadConf("", "", []string{})

				mounts := nfsv3driver.NewNfsV3ConfigDetails()
				mounts.ReadConf("sloppy_mount,allow_other,allow_root,multithread,default_permissions,fusenfs_uid,fusenfs_gid,username,password", "", []string{})

				subject = nfsv3driver.NewNfsV3Mounter(fakeInvoker, fakeOs, fakeIoutil, fakeMountChecker, nfsv3driver.NewNfsV3Config(source, mounts), fakeIdResolver)
				fakeIdResolver.ResolveReturns("100", "100", nil)

				fakeInvoker.InvokeReturns(nil, nil)
				opts["username"] = "test-user"
				opts["password"] = "test-pw"
			})

			JustBeforeEach(func() {
				err = subject.Mount(env, "source", "target", opts)
			})
			It("does not show the credentials in the options", func() {
				Expect(err).NotTo(HaveOccurred())
				_, _, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(strings.Join(args, " ")).To(Not(ContainSubstring("username")))
				Expect(strings.Join(args, " ")).To(Not(ContainSubstring("password")))
			})

			It("shows gid and uid", func() {
				Expect(err).NotTo(HaveOccurred())
				_, _, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(strings.Join(args, " ")).To(ContainSubstring("uid"))
				Expect(strings.Join(args, " ")).To(ContainSubstring("gid"))
			})

			Context("when uid and gid are passed", func() {
				BeforeEach(func() {
					opts["uid"] = "uid"
					opts["gid"] = "gid"
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Not allowed options"))
					_, ok := err.(voldriver.SafeError)
					Expect(ok).To(BeTrue())
				})
			})
			Context("when idresolver isn't present but username is passed", func() {
				BeforeEach(func() {
					source := nfsv3driver.NewNfsV3ConfigDetails()
					source.ReadConf("", "", []string{})

					mounts := nfsv3driver.NewNfsV3ConfigDetails()
					mounts.ReadConf("sloppy_mount,allow_other,allow_root,multithread,default_permissions,fusenfs_uid,fusenfs_gid,username,password", "", []string{})

					subject = nfsv3driver.NewNfsV3Mounter(fakeInvoker, fakeOs, fakeIoutil, fakeMountChecker, nfsv3driver.NewNfsV3Config(source, mounts), nil)
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("LDAP is not configured"))
					_, ok := err.(voldriver.SafeError)
					Expect(ok).To(BeTrue())
				})
			})
		})
	})

	Context("#Unmount", func() {
		Context("when mount succeeds", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns(nil, nil)

				err = subject.Unmount(env, "target")
			})

			It("should return without error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should use the passed in variables", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(Equal(1))

				_, cmd, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("umount"))
				Expect(len(args)).To(Equal(2))
				Expect(args[0]).To(Equal("-l"))
				Expect(args[1]).To(Equal("target"))
			})
		})

		Context("when unmount fails", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("error"))
				err = subject.Unmount(env, "target")
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				_, ok := err.(voldriver.SafeError)
				Expect(ok).To(BeTrue())
			})
		})
	})

	Context("#Check", func() {

		var (
			success bool
		)

		Context("when check succeeds", func() {
			BeforeEach(func() {
				success = subject.Check(env, "target", "source")
			})
			It("uses correct context", func() {
				env, _, _ := fakeInvoker.InvokeArgsForCall(0)
				Expect(fmt.Sprintf("%#v", env.Context())).To(ContainSubstring("timerCtx"))
			})
			It("reports valid mountpoint", func() {
				Expect(success).To(BeTrue())
			})
		})
		Context("when check fails", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("error"))
				success = subject.Check(env, "target", "source")
			})
			It("reports invalid mountpoint", func() {
				Expect(success).To(BeFalse())
			})
		})
	})

	Context("#CustomConfig", func() {

		var (
			sourceAllow   string
			sourceDefault string
			mountAllow    string
			mountDefault  string
		)

		Context("given allowed parameters is empty", func() {

			BeforeEach(func() {
				sourceAllow = ""
				sourceDefault = ""
				mountAllow = ""
				mountDefault = ""
			})

			Context("given allow_root=true is supplied", func() {

				BeforeEach(func() {
					source := nfsv3driver.NewNfsV3ConfigDetails()
					source.ReadConf(sourceAllow, sourceDefault, []string{})

					mounts := nfsv3driver.NewNfsV3ConfigDetails()
					mounts.ReadConf(mountAllow, mountDefault, []string{})

					subject = nfsv3driver.NewNfsV3Mounter(fakeInvoker, fakeOs, fakeIoutil, fakeMountChecker, nfsv3driver.NewNfsV3Config(source, mounts), nil)

					fakeInvoker.InvokeReturns(nil, nil)

					opts["allow_root"] = true

					err = subject.Mount(env, "source", "target", opts)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("given allowed parameters contains allow_root", func() {

			BeforeEach(func() {
				sourceAllow = ""
				sourceDefault = ""
				mountAllow = "allow_root"
				mountDefault = ""
			})

			Context("given allow_root=true is supplied", func() {

				BeforeEach(func() {

					source := nfsv3driver.NewNfsV3ConfigDetails()
					source.ReadConf(sourceAllow, sourceDefault, []string{})

					mounts := nfsv3driver.NewNfsV3ConfigDetails()
					mounts.ReadConf(mountAllow, mountDefault, []string{})

					subject = nfsv3driver.NewNfsV3Mounter(fakeInvoker, fakeOs, fakeIoutil, fakeMountChecker, nfsv3driver.NewNfsV3Config(source, mounts), nil)

					fakeInvoker.InvokeReturns(nil, nil)

					opts["allow_root"] = true

					err = subject.Mount(env, "source", "target", opts)
				})

				It("should return without error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("flows allow_root=true option through", func() {
					_, cmd, args := fakeInvoker.InvokeArgsForCall(0)

					Expect(cmd).To(Equal("fuse-nfs"))
					Expect(strings.Join(args, " ")).To(ContainSubstring("--allow_root"))
					Expect(strings.Join(args, " ")).ToNot(ContainSubstring("allow_root=true"))
					Expect(strings.Join(args, " ")).To(ContainSubstring("-n source"))
					Expect(strings.Join(args, " ")).To(ContainSubstring("-m target"))
				})
			})
		})

		Context("given sloppy_mount is true", func() {

			BeforeEach(func() {
				sourceAllow = ""
				sourceDefault = ""
				mountAllow = ""
				mountDefault = "sloppy_mount:true"
			})

			Context("given invalid parameters", func() {

				BeforeEach(func() {

					source := nfsv3driver.NewNfsV3ConfigDetails()
					source.ReadConf(sourceAllow, sourceDefault, []string{})

					mounts := nfsv3driver.NewNfsV3ConfigDetails()
					mounts.ReadConf(mountAllow, mountDefault, []string{})

					subject = nfsv3driver.NewNfsV3Mounter(fakeInvoker, fakeOs, fakeIoutil, fakeMountChecker, nfsv3driver.NewNfsV3Config(source, mounts), nil)

					fakeInvoker.InvokeReturns(nil, nil)

					opts["allow_root"] = true

					err = subject.Mount(env, "source", "target", opts)
				})

				It("should return without error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("does not flow invalid parameters", func() {
					_, cmd, args := fakeInvoker.InvokeArgsForCall(0)

					Expect(cmd).To(Equal("fuse-nfs"))
					Expect(strings.Join(args, " ")).ToNot(ContainSubstring("--allow_root"))
					Expect(strings.Join(args, " ")).ToNot(ContainSubstring("--allow_root=true"))
					Expect(strings.Join(args, " ")).To(ContainSubstring("-n source"))
					Expect(strings.Join(args, " ")).To(ContainSubstring("-m target"))
				})
			})
		})
	})

	Context("#Purge", func() {
		Context("when Purge succeeds", func() {
			BeforeEach(func() {
				fakeMountChecker.ListReturns([]string{"/var/vcap/data/some/path/guidy-guid-guid"}, nil)
			})

			JustBeforeEach(func() {
				subject.Purge(env, "/var/vcap/data/some/path")
			})

			It("should pkill fuse-nfs", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(Equal(32))

				_, proc, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(proc).To(Equal("pkill"))
				Expect(args[0]).To(Equal("fuse-nfs"))

				_, proc, args = fakeInvoker.InvokeArgsForCall(1)
				Expect(proc).To(Equal("pgrep"))
				Expect(args[0]).To(Equal("fuse-nfs"))
			})

			Context("when there are mounts", func() {
				It("should unmount the mounts", func() {
					Expect(fakeInvoker.InvokeCallCount()).To(Equal(32))

					_, proc, args := fakeInvoker.InvokeArgsForCall(31)
					Expect(proc).To(Equal("umount"))
					Expect(len(args)).To(Equal(3))
					Expect(args[0]).To(Equal("-l"))
					Expect(args[1]).To(Equal("-f"))
					Expect(args[2]).To(Equal("/var/vcap/data/some/path/guidy-guid-guid"))
				})

				It("should remove the mountpoints", func() {
					Expect(fakeOs.RemoveCallCount()).NotTo(BeZero())

					path := fakeOs.RemoveArgsForCall(0)
					Expect(path).To(Equal("/var/vcap/data/some/path/guidy-guid-guid"))
				})
			})
		})
	})
})
