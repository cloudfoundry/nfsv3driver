package nfsv3driver_test

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/nfsdriver"
	"code.cloudfoundry.org/nfsv3driver"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/voldriver/voldriverfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("NfsV3Mounter", func() {

	var (
		logger      lager.Logger
		testContext context.Context
		env         voldriver.Env
		err         error

		fakeInvoker *voldriverfakes.FakeInvoker

		subject nfsdriver.Mounter

		opts map[string]interface{}
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("nfs-mounter")
		testContext = context.TODO()
		env = driverhttp.NewHttpDriverEnv(logger, testContext)
		opts = map[string]interface{}{}

		fakeInvoker = &voldriverfakes.FakeInvoker{}

		source := nfsv3driver.NewNfsV3ConfigDetails()
		source.ReadConf("uid,gid", "", []string{})

		mounts := nfsv3driver.NewNfsV3ConfigDetails()
		mounts.ReadConf("sloppy_mount,allow_other,allow_root,multithread,default_permissions,fusenfs_uid,fusenfs_gid", "", []string{})

		subject = nfsv3driver.NewNfsV3Mounter(fakeInvoker, nfsv3driver.NewNfsV3Config(source, mounts))
	})

	Context("#Mount", func() {
		Context("when mount succeeds", func() {
			BeforeEach(func() {
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
		})

		Context("when mount errors", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("error"))

				err = subject.Mount(env, "source", "target", opts)
			})

			It("should return without error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when mount is cancelled", func() {
			// TODO: when we pick up the lager.Context
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
				_, cmd, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("fusermount"))
				Expect(strings.Join(args, " ")).To(ContainSubstring("-u target"))
			})
		})

		Context("when unmount fails", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("error"))
				err = subject.Unmount(env, "target")
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
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

					subject = nfsv3driver.NewNfsV3Mounter(fakeInvoker, nfsv3driver.NewNfsV3Config(source, mounts))

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

					subject = nfsv3driver.NewNfsV3Mounter(fakeInvoker, nfsv3driver.NewNfsV3Config(source, mounts))

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

			Context("given violation parameters", func() {

				BeforeEach(func() {

					source := nfsv3driver.NewNfsV3ConfigDetails()
					source.ReadConf(sourceAllow, sourceDefault, []string{})

					mounts := nfsv3driver.NewNfsV3ConfigDetails()
					mounts.ReadConf(mountAllow, mountDefault, []string{})

					subject = nfsv3driver.NewNfsV3Mounter(fakeInvoker, nfsv3driver.NewNfsV3Config(source, mounts))

					fakeInvoker.InvokeReturns(nil, nil)

					opts["allow_root"] = true

					err = subject.Mount(env, "source", "target", opts)
				})

				It("should return without error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("does not flow violation parameters", func() {
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
})
