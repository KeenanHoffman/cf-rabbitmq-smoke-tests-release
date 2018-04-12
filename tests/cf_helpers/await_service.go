package cf_helpers

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/types"
)

const (
	FiveSecondTimeout   time.Duration = time.Second * 5
	ThirtySecondTimeout time.Duration = time.Second * 30
	FiveMinuteTimeout   time.Duration = time.Minute * 5
	TenMinuteTimeout    time.Duration = time.Minute * 10
	ThirtyMinuteTimeout time.Duration = time.Minute * 30 // This is only so long to support a stressed director. It should be combined with smart fail-fast
)

func cfService(serviceName string) func() *gexec.Session {
	return func() *gexec.Session {
		return cf.Cf("service", serviceName)
	}
}

func cfServices() func() *gexec.Session {
	return func() *gexec.Session {
		return cf.Cf("services")
	}
}

func AwaitServiceCreation(serviceName string) {
	awaitServiceOperation(cfService(serviceName), ContainSubstring("create succeeded"))
}

func AwaitServiceDeletion(serviceName string) {
	awaitServicesOperation(serviceName, Not(ContainSubstring(serviceName)))
}

func AwaitServiceUpdate(serviceName string) {
	awaitServiceOperation(cfService(serviceName), ContainSubstring("update succeeded"))
}

func awaitServicesOperation(serviceName string, successMessageMatcher types.GomegaMatcher) {
	cfCommand := cfServices()

	Eventually(func() bool {
		session := cfCommand()
		Eventually(session, TenMinuteTimeout).Should(gexec.Exit(0))

		contents := session.Buffer().Contents()

		match, err := successMessageMatcher.Match(contents)
		if err != nil {
			Fail(err.Error())
		}

		if match {
			return true
		}

		lines := strings.Split(string(contents), "\n")
		for _, line := range lines {
			if strings.Contains(line, serviceName) && strings.Contains(line, "failed") {
				Fail(fmt.Sprintf("cf operation on service instance '%s' failed:\n"+string(contents), serviceName))
			}
		}

		time.Sleep(FiveSecondTimeout)
		return false
	}, ThirtyMinuteTimeout).Should(BeTrue())
}

func awaitServiceOperation(cfCommand func() *gexec.Session, successMessageMatcher types.GomegaMatcher) {
	Eventually(func() bool {
		session := cfCommand()
		Eventually(session, FiveMinuteTimeout).Should(gexec.Exit(0))

		contents := session.Buffer().Contents()

		match, err := successMessageMatcher.Match(contents)
		if err != nil {
			Fail(err.Error())
		}

		if match {
			return true
		}

		if bytes.Contains(contents, []byte("failed")) {
			Fail("cf operation failed:\n" + string(contents))
		}

		time.Sleep(FiveSecondTimeout)
		return false
	}, ThirtyMinuteTimeout).Should(BeTrue())

}
