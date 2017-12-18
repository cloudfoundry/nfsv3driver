package nfsv3driver_test

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/goshims/ioutilshim/ioutil_fake"
	"code.cloudfoundry.org/goshims/osshim/os_fake"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/nfsdriver"
	"code.cloudfoundry.org/nfsdriver/nfsdriverfakes"
	"code.cloudfoundry.org/nfsv3driver"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/voldriver/voldriverfakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("MapfsMounter", func() {

	var (
		logger      lager.Logger
		testContext context.Context
		env         voldriver.Env
		err         error

		fakeInvoker *voldriverfakes.FakeInvoker

		subject     nfsdriver.Mounter
		fakeMounter *nfsdriverfakes.FakeMounter
		fakeIoutil  *ioutil_fake.FakeIoutil
		fakeOs      *os_fake.FakeOs

		opts map[string]interface{}
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("mapfs-mounter")
		testContext = context.TODO()
		env = driverhttp.NewHttpDriverEnv(logger, testContext)
		opts = map[string]interface{}{}
		opts["experimental"] = true
		opts["uid"] = "2000"
		opts["gid"] = "2000"

		fakeInvoker = &voldriverfakes.FakeInvoker{}
		fakeMounter = &nfsdriverfakes.FakeMounter{}
		fakeIoutil = &ioutil_fake.FakeIoutil{}
		fakeOs = &os_fake.FakeOs{}

		fakeOs.StatReturns(nil, nil)

		subject = nfsv3driver.NewMapfsMounter(fakeInvoker, fakeMounter, fakeOs, fakeIoutil, "my-fs", "my-mount-options")
	})

	Context("#Mount", func() {
		var (
			source, target string
		)
		BeforeEach(func() {
			source = "source"
			target = "target"
		})
		JustBeforeEach(func() {
			err = subject.Mount(env, source, target, opts)
		})
		Context("when mount options don't specify experimental mounting", func() {
			BeforeEach(func() {
				delete(opts, "experimental")
			})
			It("should use the nfs_v3_mounter mounter", func() {
				Expect(fakeMounter.MountCallCount()).To(Equal(1))
				Expect(fakeInvoker.InvokeCallCount()).To(Equal(0))
			})
		})

		Context("when mount succeeds", func() {
			It("should use the mapfs mounter", func() {
				Expect(fakeMounter.MountCallCount()).To(Equal(0))
				Expect(fakeInvoker.InvokeCallCount()).NotTo(BeZero())
			})

			It("should return without error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should create an intermediary mount point", func() {
				Expect(fakeOs.MkdirAllCallCount()).NotTo(BeZero())
				dirName, mode := fakeOs.MkdirAllArgsForCall(0)
				Expect(dirName).To(Equal("target_mapfs"))
				Expect(mode).To(Equal(os.ModePerm))
			})

			It("should use the passed in variables", func() {
				_, cmd, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("mount"))
				Expect(len(args)).To(BeNumerically(">", 5))
				Expect(args[0]).To(Equal("-t"))
				Expect(args[1]).To(Equal("my-fs"))
				Expect(args[2]).To(Equal("-o"))
				Expect(args[3]).To(Equal("my-mount-options"))
				Expect(args[4]).To(Equal("source"))
				Expect(args[5]).To(Equal("target_mapfs"))
			})

			It("should launch mapfs to mount the target", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(BeNumerically(">=", 2))
				_, cmd, args := fakeInvoker.InvokeArgsForCall(1)
				Expect(cmd).To(Equal("mapfs"))
				Expect(args[0]).To(Equal("-uid"))
				Expect(args[1]).To(Equal("2000"))
				Expect(args[2]).To(Equal("-gid"))
				Expect(args[3]).To(Equal("2000"))
				Expect(args[4]).To(Equal("target"))
				Expect(args[5]).To(Equal("target_mapfs"))
			})

			Context("when the mount has a legacy format", func(){
				BeforeEach(func(){
					source = "nfs://server/some/share/path"
				})
				It("should rewrite the share to use standard nfs format", func(){
					_, _, args := fakeInvoker.InvokeArgsForCall(0)
					Expect(len(args)).To(BeNumerically(">", 4))
					Expect(args[4]).To(Equal("server:/some/share/path"))
				})
			})
		})
		Context("when there is no uid", func() {
			BeforeEach(func() {
				delete(opts, "uid")
			})
			It("should error", func() {
				Expect(err).ToNot(BeNil())
			})
		})
		Context("when there is no gid", func() {
			BeforeEach(func() {
				delete(opts, "gid")
			})
			It("should error", func() {
				Expect(err).ToNot(BeNil())
			})
		})
		Context("when uid is an integer", func() {
			BeforeEach(func() {
				opts["uid"] = 2000
			})
			It("should not error", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(BeNumerically(">=", 2))
				_, cmd, args := fakeInvoker.InvokeArgsForCall(1)
				Expect(cmd).To(Equal("mapfs"))
				Expect(args[0]).To(Equal("-uid"))
				Expect(args[1]).To(Equal("2000"))
				Expect(args[2]).To(Equal("-gid"))
				Expect(args[3]).To(Equal("2000"))
				Expect(args[4]).To(Equal("target"))
				Expect(args[5]).To(Equal("target_mapfs"))
			})
		})
		Context("when gid is an integer", func() {
			BeforeEach(func() {
				opts["gid"] = 2000
			})
			It("should not error", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(BeNumerically(">=", 2))
				_, cmd, args := fakeInvoker.InvokeArgsForCall(1)
				Expect(cmd).To(Equal("mapfs"))
				Expect(args[0]).To(Equal("-uid"))
				Expect(args[1]).To(Equal("2000"))
				Expect(args[2]).To(Equal("-gid"))
				Expect(args[3]).To(Equal("2000"))
				Expect(args[4]).To(Equal("target"))
				Expect(args[5]).To(Equal("target_mapfs"))
			})
		})

		Context("when mount errors", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("error"))

				err = subject.Mount(env, "source", "target", opts)
			})

			It("should return error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("#Unmount", func() {
		JustBeforeEach(func() {
			err = subject.Unmount(env, "target")
		})

		Context("when mount is not a mapfs mount", func() {
			BeforeEach(func() {
				fakeOs.StatReturns(nil, errors.New("badness"))
			})

			It("should use the nfs_v3_mounter mounter", func() {
				Expect(fakeMounter.UnmountCallCount()).To(Equal(1))
				Expect(fakeInvoker.InvokeCallCount()).To(Equal(0))
			})
		})

		Context("when unmount succeeds", func() {
			It("should return without error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should invoke unmount on both the mapfs and target mountpoints", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(BeNumerically(">", 1))
				_, cmd, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("umount"))
				Expect(args[0]).To(Equal("target"))
				_, cmd, args = fakeInvoker.InvokeArgsForCall(1)
				Expect(cmd).To(Equal("umount"))
				Expect(args[0]).To(Equal("target_mapfs"))
			})
			It("should delete the mapfs mount point", func() {
				Expect(fakeOs.RemoveAllCallCount()).ToNot(BeZero())
				Expect(fakeOs.RemoveAllArgsForCall(0)).To(Equal("target_mapfs"))
			})
		})

		Context("when unmount fails", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("error"))
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

	Context("#Purge", func() {
		Context("when Purge succeeds", func() {
			JustBeforeEach(func() {
				subject.Purge(env, "/foo/foo/foo")
			})
			It("kills the mapfs mount processes", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(BeNumerically(">=", 2))
				_, proc, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(proc).To(Equal("pkill"))
				Expect(args[0]).To(Equal("mapfs"))
				_, proc, args = fakeInvoker.InvokeArgsForCall(1)
				Expect(proc).To(Equal("pgrep"))
				Expect(args[0]).To(Equal("mapfs"))
			})
			Context("when there are mapfs mounts", func() {
				BeforeEach(func() {
					fakeMapfsDir := &ioutil_fake.FakeFileInfo{}
					fakeMapfsDir.NameReturns("mount_one_mapfs")
					fakeMapfsDir.IsDirReturns(true)

					fakeIoutil.ReadDirReturns([]os.FileInfo{fakeMapfsDir}, nil)
				})
				It("should unmount the mapfs mount", func() {
					Expect(fakeInvoker.InvokeCallCount()).To(BeNumerically(">=", 1))
					_, cmd, args := fakeInvoker.InvokeArgsForCall(fakeInvoker.InvokeCallCount() - 1)
					Expect(cmd).To(Equal("umount"))
					Expect(args[0]).To(Equal("-f"))
					Expect(args[1]).To(Equal("/foo/foo/foo/mount_one"))
				})
				It("should remove both the mountpoints", func() {
					Expect(fakeOs.RemoveAllCallCount()).To(BeNumerically(">=", 2))
					path := fakeOs.RemoveAllArgsForCall(0)
					Expect(path).To(Equal("/foo/foo/foo/mount_one"))
					path = fakeOs.RemoveAllArgsForCall(1)
					Expect(path).To(Equal("/foo/foo/foo/mount_one_mapfs"))
				})
			})
			It("eventually calls purge on the v3 mounter", func() {
				Expect(fakeMounter.PurgeCallCount()).To(Equal(1))
			})
		})

	})
})
