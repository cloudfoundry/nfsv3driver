package nfsv3driver_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/onsi/gomega/gbytes"
	"os"
	"strings"
	"syscall"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/dockerdriverfakes"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/goshims/ioutilshim/ioutil_fake"
	"code.cloudfoundry.org/goshims/osshim/os_fake"
	"code.cloudfoundry.org/goshims/syscallshim/syscall_fake"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/nfsv3driver"
	"code.cloudfoundry.org/nfsv3driver/nfsdriverfakes"
	vmo "code.cloudfoundry.org/volume-mount-options"
	"code.cloudfoundry.org/volumedriver"
	nfsfakes "code.cloudfoundry.org/volumedriver/volumedriverfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MapfsMounter", func() {

	var (
		logger      *lagertest.TestLogger
		testContext context.Context
		env         dockerdriver.Env
		err         error

		fakePgInvoker  *dockerdriverfakes.FakeInvoker
		fakeInvoker    *dockerdriverfakes.FakeInvoker
		fakeBgInvoker  *nfsdriverfakes.FakeBackgroundInvoker
		fakeIdResolver *nfsdriverfakes.FakeIdResolver

		subject          volumedriver.Mounter
		fakeIoutil       *ioutil_fake.FakeIoutil
		fakeOs           *os_fake.FakeOs
		fakeMountChecker *nfsfakes.FakeMountChecker
		fakeSyscall      *syscall_fake.FakeSyscall

		opts      map[string]interface{}
		mapfsPath string
		mask      vmo.MountOptsMask
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("mapfs-mounter")
		mapfsPath = "/var/vcap/packages/mapfs/bin/mapfs"
		testContext = context.TODO()
		env = driverhttp.NewHttpDriverEnv(logger, testContext)
		opts = map[string]interface{}{}
		opts["uid"] = "2000"
		opts["gid"] = "2000"

		fakePgInvoker = &dockerdriverfakes.FakeInvoker{}
		fakeInvoker = &dockerdriverfakes.FakeInvoker{}
		fakeBgInvoker = &nfsdriverfakes.FakeBackgroundInvoker{}
		fakeIoutil = &ioutil_fake.FakeIoutil{}
		fakeOs = &os_fake.FakeOs{}
		fakeSyscall = &syscall_fake.FakeSyscall{}
		fakeOs.OpenFileReturns(&os_fake.FakeFile{}, nil)
		fakeMountChecker = &nfsfakes.FakeMountChecker{}
		fakeMountChecker.ExistsReturns(true, nil)

		fakeOs.StatReturns(nil, nil)
		fakeOs.IsExistReturns(true)

		fakeSyscall.StatStub = func(path string, st *syscall.Stat_t) error {
			st.Mode = 0777
			st.Uid = 1000
			st.Gid = 1000
			return nil
		}

		mask, err = nfsv3driver.NewMapFsVolumeMountMask("auto_cache,fsname", "")
		Expect(err).NotTo(HaveOccurred())

		subject = nfsv3driver.NewMapfsMounter(fakePgInvoker, fakeInvoker, fakeBgInvoker, fakeOs, fakeSyscall, fakeIoutil, fakeMountChecker, "my-fs", "my-mount-options,timeo=600,retrans=2,actimeo=0", nil, mask, mapfsPath)
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

		Context("when version is specified", func() {
			BeforeEach(func() {
				opts["version"] = "4.1"
			})

			It("should use version specified", func() {
				Expect(err).NotTo(HaveOccurred())
				_, cmd, args := fakePgInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("mount"))
				Expect(len(args)).To(BeNumerically(">", 5))
				Expect(args).To(ContainElement("-t"))
				Expect(args).To(ContainElement("my-fs"))
				Expect(args).To(ContainElement("-o"))
				Expect(args).To(ContainElement("my-mount-options,timeo=600,retrans=2,actimeo=0,vers=4.1"))
				Expect(args).To(ContainElement("source"))
				Expect(args).To(ContainElement("target_mapfs"))
			})

		})

		Context("when mount succeeds", func() {
			It("should use the mapfs mounter", func() {
				Expect(fakePgInvoker.InvokeCallCount()).NotTo(BeZero())
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
				_, cmd, args := fakePgInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("mount"))
				Expect(len(args)).To(BeNumerically(">", 5))
				Expect(args).To(ContainElement("-t"))
				Expect(args).To(ContainElement("my-fs"))
				Expect(args).To(ContainElement("-o"))
				Expect(args).To(ContainElement("my-mount-options,timeo=600,retrans=2,actimeo=0"))
				Expect(args).To(ContainElement("source"))
				Expect(args).To(ContainElement("target_mapfs"))
			})

			It("should launch mapfs to mount the target", func() {
				Expect(fakeBgInvoker.InvokeCallCount()).To(BeNumerically(">=", 1))
				_, cmd, args, waitFor, _ := fakeBgInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal(mapfsPath))
				Expect(args).To(ContainElement("-uid"))
				Expect(args).To(ContainElement("2000"))
				Expect(args).To(ContainElement("-gid"))
				Expect(args).To(ContainElement("target"))
				Expect(args).To(ContainElement("target_mapfs"))
				Expect(waitFor).To(Equal("Mounted!"))
			})

			Context("when mkdir fails", func() {
				BeforeEach(func() {
					fakeOs.MkdirAllReturns(errors.New("failed-to-create-dir"))
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					_, ok := err.(dockerdriver.SafeError)
					Expect(ok).To(BeTrue())
				})
			})

			Context("when the mount is readonly", func() {
				BeforeEach(func() {
					opts["readonly"] = true
				})

				It("should not append 'ro' to the kernel mount options, since garden manages the ro mount", func() {
					_, _, args := fakePgInvoker.InvokeArgsForCall(0)
					Expect(len(args)).To(BeNumerically(">", 3))
					Expect(args[2]).To(Equal("-o"))
					Expect(args[3]).NotTo(ContainSubstring(",ro"))
				})
				It("should not append 'actimeo=0' to the kernel mount options", func() {
					Expect(err).NotTo(HaveOccurred())
					_, _, args := fakePgInvoker.InvokeArgsForCall(0)
					Expect(len(args)).To(BeNumerically(">", 3))
					Expect(args[2]).To(Equal("-o"))
					Expect(args[3]).NotTo(ContainSubstring("actimeo=0"))
				})
			})

			Context("when the mount has a legacy format", func() {
				BeforeEach(func() {
					source = "nfs://server/some/share/path"
				})
				It("should rewrite the share to use standard nfs format", func() {
					_, _, args := fakePgInvoker.InvokeArgsForCall(0)
					Expect(len(args)).To(BeNumerically(">", 4))
					Expect(args[4]).To(Equal("server:/some/share/path"))
				})
			})

			Context("when the mount has a legacy format without subdirectory", func() {
				BeforeEach(func() {
					source = "nfs://server/"
				})
				It("should rewrite the share to use standard nfs format", func() {
					_, _, args := fakePgInvoker.InvokeArgsForCall(0)
					Expect(len(args)).To(BeNumerically(">", 4))
					Expect(args[4]).To(Equal("server:/"))
				})
			})

			Context("when the target has a trailing slash", func() {
				BeforeEach(func() {
					target = "/some/target/"
				})
				It("should rewrite the target to remove the slash", func() {
					Expect(fakeBgInvoker.InvokeCallCount()).To(BeNumerically(">=", 1))
					_, _, args, _, _ := fakeBgInvoker.InvokeArgsForCall(0)
					Expect(args[4]).To(Equal("/some/target"))
					Expect(args[5]).To(Equal("/some/target_mapfs"))
				})
			})

			Context("when other options are specified", func() {
				BeforeEach(func() {
					opts["auto_cache"] = true
					opts["fsname"] = "zanzibar"
				})
				It("should include those options on the mapfs invoke call", func() {
					Expect(fakeBgInvoker.InvokeCallCount()).To(BeNumerically(">=", 1))
					_, _, args, _, _ := fakeBgInvoker.InvokeArgsForCall(0)
					Expect(args).To(ContainElement("-auto_cache"))
					Expect(args).To(ContainElement("-fsname"))
					Expect(args).To(ContainElement("zanzibar"))
				})
			})
		})
		Context("when there is no uid", func() {
			BeforeEach(func() {
				delete(opts, "uid")
			})

			It("should create an intermediary mount point", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeOs.MkdirAllCallCount()).NotTo(BeZero())
				dirName, _ := fakeOs.MkdirAllArgsForCall(0)
				Expect(dirName).To(Equal("target_mapfs"))
			})

			It("should mount directly to the target", func() {
				Expect(err).NotTo(HaveOccurred())
				_, cmd, args := fakePgInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("mount"))
				Expect(args).To(ContainElement("source"))
				Expect(args).To(ContainElement("target"))
			})

			It("should not launch mapfs", func() {
				Expect(fakeBgInvoker.InvokeCallCount()).To(Equal(0))
			})
		})
		Context("when there is no gid", func() {
			BeforeEach(func() {
				delete(opts, "gid")
			})
			It("should error", func() {
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
			})
		})
		Context("when uid is an integer", func() {
			BeforeEach(func() {
				opts["uid"] = 2000
			})
			It("should not error", func() {
				Expect(fakeBgInvoker.InvokeCallCount()).To(BeNumerically(">=", 1))
				_, cmd, args, _, _ := fakeBgInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal(mapfsPath))
				Expect(args).To(ContainElement("-uid"))
				Expect(args).To(ContainElement("2000"))
				Expect(args).To(ContainElement("-gid"))
				Expect(args).To(ContainElement("target"))
				Expect(args).To(ContainElement("target_mapfs"))
			})
		})
		Context("when gid is an integer", func() {
			BeforeEach(func() {
				opts["gid"] = 2000
			})
			It("should not error", func() {
				Expect(fakeBgInvoker.InvokeCallCount()).To(BeNumerically(">=", 1))
				_, cmd, args, _, _ := fakeBgInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal(mapfsPath))
				Expect(args).To(ContainElement("-uid"))
				Expect(args).To(ContainElement("2000"))
				Expect(args).To(ContainElement("-gid"))
				Expect(args).To(ContainElement("target"))
				Expect(args).To(ContainElement("target_mapfs"))
			})
		})
		Context("when the specified uid doesn't have read access", func() {
			BeforeEach(func() {
				fakeSyscall.StatStub = func(path string, st *syscall.Stat_t) error {
					st.Mode = 0750
					st.Uid = 1000
					st.Gid = 1000
					return nil
				}
			})
			It("should fail and clean up the intermediate mount", func() {
				Expect(fakeSyscall.StatCallCount()).NotTo(BeZero())
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
				Expect(err.Error()).To(ContainSubstring("access"))

				Expect(fakeInvoker.InvokeCallCount()).To(Equal(1))
				_, cmd, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("umount"))
				Expect(len(args)).To(BeNumerically(">", 0))
				Expect(args[0]).To(Equal("target_mapfs"))
				Expect(fakeOs.RemoveCallCount()).To(Equal(1))

				Expect(logger.LogMessages()).NotTo(ContainElement(ContainSubstring("intermediate-unmount-failed")))
				Expect(logger.LogMessages()).NotTo(ContainElement(ContainSubstring("intermediate-remove-failed")))
			})

			Context("when it fails to unmount the intermediate directory", func() {

				BeforeEach(func() {
					fakeInvoker.InvokeReturns(nil, errors.New("intermediate-unmount-failed"))
				})
				It("should log the error", func() {
					Expect(logger.LogMessages()).To(ContainElement(ContainSubstring("intermediate-unmount-failed")))
					Expect(fakeOs.RemoveCallCount()).To(Equal(0))
				})
			})

			Context("when it fails to remove the intermediate directory", func() {

				BeforeEach(func() {
					fakeOs.RemoveReturns(errors.New("intermediate-remove-failed"))
				})
				It("should log the error", func() {
					Expect(logger.LogMessages()).To(ContainElement(ContainSubstring("intermediate-remove-failed")))
				})
			})
		})
		Context("when stat() fails during access check", func() {
			BeforeEach(func() {
				fakeSyscall.StatStub = func(path string, st *syscall.Stat_t) error {
					return errors.New("this is nacho share.")
				}
			})
			It("should succeed and log a warning", func() {
				Expect(fakeSyscall.StatCallCount()).NotTo(BeZero())
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.TestSink.Buffer().Contents()).To(ContainSubstring("nacho share"))
			})


		})
		Context("when stat returns ambiguous results", func() {
			var (
				uid = uint32(1000)
				gid = uint32(1000)
			)
			BeforeEach(func() {
				fakeSyscall.StatStub = func(path string, st *syscall.Stat_t) error {
					st.Mode = 0750
					st.Uid = uid
					st.Gid = gid
					return nil
				}
			})

			Context("when uid is unknown", func() {
				BeforeEach(func() {
					uid = nfsv3driver.UnknownId
				})
				It("should succeed", func() {
					Expect(fakeSyscall.StatCallCount()).NotTo(BeZero())
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when gid is unknown", func() {
				BeforeEach(func() {
					gid = nfsv3driver.UnknownId
				})
				It("should succeed", func() {
					Expect(fakeSyscall.StatCallCount()).NotTo(BeZero())
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
		Context("when idresolver isn't present but username is passed", func() {
			BeforeEach(func() {
				delete(opts, "uid")
				delete(opts, "gid")
				opts["username"] = "test-user"
				opts["password"] = "test-pw"
			})

			It("should error", func() {
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
				Expect(err.Error()).To(ContainSubstring("LDAP is not configured"))
			})
		})
		Context("when mount errors", func() {
			BeforeEach(func() {
				fakePgInvoker.InvokeReturns([]byte("error"), fmt.Errorf("error"))
			})

			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
			})

			It("should remove the intermediary mountpoint", func() {
				Expect(logger.LogMessages()).NotTo(ContainElement(ContainSubstring("remove-failed")))

				Expect(fakeOs.RemoveCallCount()).To(Equal(1))
			})

			Context("when the intermediate mount directory remove fails", func() {
				BeforeEach(func() {
					fakeOs.RemoveReturns(errors.New("remove-failed"))
				})

				It("should log an error", func() {
					Expect(logger.LogMessages()).To(ContainElement(ContainSubstring("remove-failed")))
				})
			})
		})
		Context("when kernel mount succeeds, but mapfs mount fails", func() {
			BeforeEach(func() {
				fakeBgInvoker.InvokeReturns(fmt.Errorf("error"), nil)
			})

			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
			})
			It("should invoke unmount", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(Equal(1))
				_, cmd, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("umount"))
				Expect(len(args)).To(BeNumerically(">", 0))
				Expect(args[0]).To(Equal("target_mapfs"))
			})
			It("should remove the intermediary mountpoint", func() {
				Expect(fakeOs.RemoveCallCount()).To(Equal(1))
				Expect(logger.LogMessages()).NotTo(ContainElement(ContainSubstring("unmount-failed")))
				Expect(logger.LogMessages()).NotTo(ContainElement(ContainSubstring("remove-failed")))
			})

			Context("when unmount fails", func() {
				BeforeEach(func(){
					fakeInvoker.InvokeReturns(nil, errors.New("unmount-failed"))
				})
				It("should log the error", func() {
					Expect(logger.LogMessages()).To(ContainElement(ContainSubstring("unmount-failed")))
					Expect(fakeOs.RemoveCallCount()).To(Equal(0))
				})
			})

			Context("when remove fails", func() {
				BeforeEach(func(){
					fakeOs.RemoveReturns(errors.New("remove-failed"))
				})
				It("should log the error", func() {
					Expect(logger.LogMessages()).To(ContainElement(ContainSubstring("remove-failed")))
				})
			})
		})

		Context("when provided a username to map to a uid", func() {
			BeforeEach(func() {
				fakeIdResolver = &nfsdriverfakes.FakeIdResolver{}

				subject = nfsv3driver.NewMapfsMounter(fakePgInvoker, fakeInvoker, fakeBgInvoker, fakeOs, fakeSyscall, fakeIoutil, fakeMountChecker, "my-fs", "my-mount-options", fakeIdResolver, mask, mapfsPath)
				fakeIdResolver.ResolveReturns("100", "100", nil)

				delete(opts, "uid")
				delete(opts, "gid")
				opts["username"] = "test-user"
				opts["password"] = "test-pw"
			})

			It("does not show the credentials in the options", func() {
				Expect(err).NotTo(HaveOccurred())
				_, _, args, _, _ := fakeBgInvoker.InvokeArgsForCall(0)
				Expect(strings.Join(args, " ")).To(Not(ContainSubstring("username")))
				Expect(strings.Join(args, " ")).To(Not(ContainSubstring("password")))
			})

			It("shows gid and uid", func() {
				Expect(err).NotTo(HaveOccurred())
				_, _, args, _, _ := fakeBgInvoker.InvokeArgsForCall(0)
				Expect(strings.Join(args, " ")).To(ContainSubstring("-uid 100"))
				Expect(strings.Join(args, " ")).To(ContainSubstring("-gid 100"))
			})

			Context("when username is passed but password is not passed", func() {
				BeforeEach(func() {
					delete(opts, "password")
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
					_, ok := err.(dockerdriver.SafeError)
					Expect(ok).To(BeTrue())
					Expect(err.Error()).To(ContainSubstring("LDAP password is missing"))
				})
			})

			Context("when uid is NaN", func() {
				BeforeEach(func() {
					fakeIdResolver.ResolveReturns("uid-not-a-number", "1", nil)
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("uid-not-a-number"))
				})
			})

			Context("when gid is NaN", func() {
				BeforeEach(func() {
					fakeIdResolver.ResolveReturns("1", "gid-not-a-number", nil)
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("gid-not-a-number"))
				})
			})

			Context("when uid is passed", func() {
				BeforeEach(func() {
					opts["uid"] = "100"
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
					_, ok := err.(dockerdriver.SafeError)
					Expect(ok).To(BeTrue())
					Expect(err.Error()).To(ContainSubstring("Not allowed options"))
				})
			})

			Context("when gid is passed", func() {
				BeforeEach(func() {
					opts["gid"] = "100"
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
					_, ok := err.(dockerdriver.SafeError)
					Expect(ok).To(BeTrue())
					Expect(err.Error()).To(ContainSubstring("Not allowed options"))
				})
			})

			Context("when unable to resolve username", func() {
				BeforeEach(func() {
					fakeIdResolver.ResolveReturns("", "", errors.New("unable to resolve"))
				})

				It("return an error that is not a SafeError since it might contain sensitive information", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("unable to resolve"))
					Expect(err).NotTo(BeAssignableToTypeOf(dockerdriver.SafeError{}))
				})
			})
		})
	})

	Context("#Unmount", func() {
		var target string
		BeforeEach(func() {
			target = "target"
		})

		JustBeforeEach(func() {
			err = subject.Unmount(env, target)
		})

		Context("when unmount succeeds", func() {
			It("should return without error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should invoke unmount on both the mapfs and target mountpoints", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(Equal(2))

				_, cmd, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("umount"))
				Expect(len(args)).To(Equal(2))
				Expect(args[0]).To(Equal("-l"))
				Expect(args[1]).To(Equal("target"))

				_, cmd, args = fakeInvoker.InvokeArgsForCall(1)
				Expect(cmd).To(Equal("umount"))
				Expect(len(args)).To(Equal(2))
				Expect(args[0]).To(Equal("-l"))
				Expect(args[1]).To(Equal("target_mapfs"))
			})

			It("should delete the mapfs mount point", func() {
				Expect(fakeOs.RemoveCallCount()).ToNot(BeZero())
				Expect(fakeOs.RemoveArgsForCall(0)).To(Equal("target_mapfs"))
			})

			Context("when the target has a trailing slash", func() {
				BeforeEach(func() {
					target = "/some/target/"
				})

				It("should rewrite the target to remove the slash", func() {
					Expect(fakeInvoker.InvokeCallCount()).To(Equal(2))

					_, cmd, args := fakeInvoker.InvokeArgsForCall(0)
					Expect(cmd).To(Equal("umount"))
					Expect(len(args)).To(Equal(2))
					Expect(args[0]).To(Equal("-l"))
					Expect(args[1]).To(Equal("/some/target"))

					_, cmd, args = fakeInvoker.InvokeArgsForCall(1)
					Expect(cmd).To(Equal("umount"))
					Expect(len(args)).To(Equal(2))
					Expect(args[0]).To(Equal("-l"))
					Expect(args[1]).To(Equal("/some/target_mapfs"))
				})
			})

			Context("when uid mapping was not used for the mount", func() {
				BeforeEach(func() {
					fakeMountChecker.ExistsStub = func(s string) (bool, error) {
						Expect(s).To(Equal("target_mapfs"))
						return false, nil
					}
				})

				It("should not attempt to unmount the intermediate mount", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeInvoker.InvokeCallCount()).To(Equal(1))
				})

				Context("when the mount checker returns an error", func() {
					BeforeEach(func() {
						fakeMountChecker.ExistsReturns(false, errors.New("mount-checker-failed"))
					})

					It("should log the error", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(logger.Buffer()).To(gbytes.Say("mount-checker-failed"))
					})
				})
			})

			Context("when the intermediate directory does not exist", func() {
				BeforeEach(func() {
					fakeOs.StatStub = func(name string) (os.FileInfo, error) {
						Expect(name).To(Equal("target_mapfs"))
						return nil, &os.PathError{Err: os.ErrNotExist}
					}
				})

				It("should succeeed", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeOs.RemoveCallCount()).To(Equal(0))
				})
			})
		})

		Context("when unmount fails", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("error"))
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
			})
		})

		Context("when unmount of the intermediate mount fails", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturnsOnCall(0, nil, nil)
				fakeInvoker.InvokeReturnsOnCall(1, []byte("error"), fmt.Errorf("mapfs umount error"))
			})

			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.Buffer()).To(gbytes.Say("mapfs umount error"))
			})

			It("should not call Remove on the intermediate directory", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeOs.RemoveCallCount()).To(Equal(0))
			})
		})

		Context("when remove fails", func() {
			BeforeEach(func() {
				fakeOs.RemoveReturns(errors.New("failed-to-remove-dir"))
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
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

	Context("#Purge", func() {
		var pathToPurge string

		BeforeEach(func() {
			pathToPurge = "/foo/foo/foo"
			fakeMountChecker.ListReturns([]string{"/foo/foo/foo/mount_one_mapfs"}, nil)
		})

		JustBeforeEach(func() {
			subject.Purge(env, pathToPurge)
		})

		Context("when Purge succeeds", func() {
			It("kills the mapfs mount processes", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(Equal(33))

				_, proc, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(proc).To(Equal("pkill"))
				Expect(args[0]).To(Equal("mapfs"))

				_, proc, args = fakeInvoker.InvokeArgsForCall(1)
				Expect(proc).To(Equal("pgrep"))
				Expect(args[0]).To(Equal("mapfs"))
			})

			Context("when there are mapfs mounts", func() {
				It("should unmount both the mounts", func() {
					Expect(fakeInvoker.InvokeCallCount()).To(Equal(33))

					_, cmd, args := fakeInvoker.InvokeArgsForCall(fakeInvoker.InvokeCallCount() - 2)
					Expect(cmd).To(Equal("umount"))
					Expect(len(args)).To(Equal(3))
					Expect(args[0]).To(Equal("-l"))
					Expect(args[1]).To(Equal("-f"))
					Expect(args[2]).To(Equal("/foo/foo/foo/mount_one"))

					_, cmd, args = fakeInvoker.InvokeArgsForCall(fakeInvoker.InvokeCallCount() - 1)
					Expect(cmd).To(Equal("umount"))
					Expect(len(args)).To(Equal(3))
					Expect(args[0]).To(Equal("-l"))
					Expect(args[1]).To(Equal("-f"))
					Expect(args[2]).To(Equal("/foo/foo/foo/mount_one_mapfs"))
				})

				It("should remove both the mountpoints", func() {
					Expect(fakeOs.RemoveCallCount()).To(Equal(2))

					path := fakeOs.RemoveArgsForCall(0)
					Expect(path).To(Equal("/foo/foo/foo/mount_one"))

					path = fakeOs.RemoveArgsForCall(1)
					Expect(path).To(Equal("/foo/foo/foo/mount_one_mapfs"))
				})
			})

			Context("when given a path to purge that is a malformed URI", func() {
				BeforeEach(func() {
					pathToPurge = "foo("
				})

				It("should log an error", func() {
					Expect(logger.TestSink.Buffer()).Should(gbytes.Say("unable-to-list-mounts"))
					Expect(fakeMountChecker.ListCallCount()).To(Equal(0))
				})
			})
		})

		Context("when unable to list mounts", func() {
			BeforeEach(func() {
				fakeMountChecker.ListReturns(nil, errors.New("list-failed"))
			})

			It("should log the error and not attempt any unmounts", func() {
				Expect(logger.Buffer()).To(gbytes.Say("list-failed"))
				Expect(logger.Buffer()).NotTo(gbytes.Say("mount-directory-list"))

			})
		})
	})

	Context("NewMapFsVolumeMountMask", func() {

		Context("when given additional options", func() {
			var (
				mask                                 vmo.MountOptsMask
				err                                  error
				allowMountOption, defaultMountOption string
			)

			BeforeEach(func() {
				allowMountOption = "opt1,opt2"
				defaultMountOption = "opt1:val1,opt2:val2"
			})

			JustBeforeEach(func() {
				mask, err = nfsv3driver.NewMapFsVolumeMountMask(allowMountOption, defaultMountOption)
			})

			It("should create a mask with those options", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mask.Allowed).To(ContainElement("opt1"))
				Expect(mask.Allowed).To(ContainElement("opt2"))
				Expect(mask.Defaults).To(HaveKeyWithValue("opt1", "val1"))
				Expect(mask.Defaults).To(HaveKeyWithValue("opt2", "val2"))
			})
		})
	})
})
