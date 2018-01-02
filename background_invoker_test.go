package nfsv3driver_test

import (
	"code.cloudfoundry.org/goshims/execshim/exec_fake"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/nfsv3driver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"bytes"
	"code.cloudfoundry.org/goshims/execshim"
	"time"
)

var _ = Describe("Background Invoker", func() {
	var (
		subject    nfsv3driver.BackgroundInvoker
		fakeCmd    *exec_fake.FakeCmd
		fakeExec   *exec_fake.FakeExec
		testLogger lager.Logger
		testEnv    voldriver.Env
		cmd        = "some-fake-command"
		args       = []string{"fake-args-1"}
		timeout = time.Millisecond * 500
	)
	Context("when invoking an executable", func() {
		BeforeEach(func() {
			testLogger = lagertest.NewTestLogger("InvokerTest")
			testEnv = driverhttp.NewHttpDriverEnv(testLogger, nil)
			fakeExec = new(exec_fake.FakeExec)
			fakeCmd = new(exec_fake.FakeCmd)
			fakeExec.CommandContextReturns(fakeCmd)
			fakeCmd.StdoutPipeReturns(&nopCloser{bytes.NewBufferString("foo\nfoo\nMounted!\nfoo\n")}, nil)

			subject = nfsv3driver.NewBackgroundInvoker(fakeExec)
		})

		It("should successfully invoke cli", func() {
			err := subject.Invoke(testEnv, cmd, args, "Mounted!", timeout)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeExec.CommandContextCallCount()).To(Equal(1))
		})

		Context("when command exits without emitting waitFor", func() {
			BeforeEach(func() {
				fakeCmd.StdoutPipeReturns(&nopCloser{bytes.NewBufferString("foo\nfoo\nfoo\n")}, nil)
			})

			It("should report an error", func() {
				err := subject.Invoke(testEnv, cmd, args, "Mounted!", timeout)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("command exited"))
			})

			Context("when we aren't waiting for anything", func(){
				It("should successfully invoke cli", func() {
					err := subject.Invoke(testEnv, cmd, args, "", timeout)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when the command takes too long to finish", func(){
			BeforeEach(func(){
				// use a real invoker for this test so that we can sleep
				subject = nfsv3driver.NewBackgroundInvoker(&execshim.ExecShim{})
				cmd        = "sleep"
				args       = []string{"15"}
			})

			It("should report an error", func() {
				err := subject.Invoke(testEnv, cmd, args, "Mounted!", timeout)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("command timed out"))
			})
		})
	})
})

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }
