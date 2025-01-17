package smoke_tests

import (
	"fmt"

	"rabbitmq-smoke-tests/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
)

var _ = Describe("Smoke tests", func() {

	const appName = "rmq-smoke-tests-ruby"
	const appPath = "../assets/rabbit-example-app"

	AfterEach(func() {
		helper.PrintAppLogs(appName)
		helper.DeleteApp(appName)
	})

	smokeTestForPlan := func(planName string, createServiceWithTLS bool) func() {
		return func() {
			serviceName := fmt.Sprintf("rmq-smoke-test-instance-%s", uuid.New()[:18])
			serviceKeyName := fmt.Sprintf("%s-key", serviceName)

			if testConfig.ServiceOffering == "p.rabbitmq" && testConfig.BindingWithDNS {
				By("creating the service instance with TLS enabled")
				tlsString := fmt.Sprintf(`{"tls": %v}`, createServiceWithTLS)
				helper.CreateService(testConfig.ServiceOffering, planName, serviceName, tlsString)
			} else {
				By("creating the service instance")
				helper.CreateService(testConfig.ServiceOffering, planName, serviceName, "")
			}

			helper.CreateServiceKey(serviceName, serviceKeyName)

			defer func() {
				By("deleting the service key")
				helper.DeleteServiceKey(serviceName, serviceKeyName)
				By("deleting the service instance")
				helper.DeleteService(serviceName)
			}()

			if createServiceWithTLS && testConfig.ServiceOffering == "p.rabbitmq" && !testConfig.BindingWithDNS {
				By("updating the service to enable TLS")
				tlsConfig := helper.TLSConfigUsingIPs(serviceName)
				helper.UpdateService(serviceName, tlsConfig)
			}

			defer func() {
				By("unbinding the app")
				helper.UnbindService(appName, serviceName)
			}()

			By("pushing and binding an app")
			appURL := helper.PushAndBindApp(appName, serviceName, appPath, testConfig.AppsDomain)

			By("sending and receiving rabbit messages")
			queue := fmt.Sprintf("%s-queue", appName)

			helper.SendMessage(appURL, queue, "foo")
			helper.SendMessage(appURL, queue, "bar")
			Expect(helper.ReceiveMessage(appURL, queue)).To(Equal("foo"))
			Expect(helper.ReceiveMessage(appURL, queue)).To(Equal("bar"))

			// By("accessing the management dashboard")
			// serviceKey := helper.GetServiceKey(serviceName, serviceKeyName)
			//
			// client := http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
			// resp, err := client.Get(serviceKey.DashboardUrl)
			// Expect(err).NotTo(HaveOccurred())
			//
			// defer resp.Body.Close()
			// Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}
	}

	switch testConfig.TLSSupport {
	case "disabled":
		for _, plan := range testConfig.TestPlans {
			It(fmt.Sprintf("pushes an app, sends, and reads a message from RabbitMQ: plan '%s'", plan.Name),
				smokeTestForPlan(plan.Name, false))
		}
	case "optional":
		for _, plan := range testConfig.TestPlans {
			It(fmt.Sprintf("pushes an app, sends, and reads a message from RabbitMQ: plan '%s'", plan.Name),
				smokeTestForPlan(plan.Name, false))
			It(fmt.Sprintf("pushes an app, sends, and reads a message from RabbitMQ over TLS: plan '%s'", plan.Name),
				smokeTestForPlan(plan.Name, true))
		}
	case "enforced":
		for _, plan := range testConfig.TestPlans {
			It(fmt.Sprintf("pushes an app, sends, and reads a message from RabbitMQ over TLS: plan '%s'", plan.Name),
				smokeTestForPlan(plan.Name, true))
		}
	default:
		for _, plan := range testConfig.TestPlans {
			It(fmt.Sprintf("pushes an app, sends, and reads a message from RabbitMQ: plan '%s'", plan.Name),
				smokeTestForPlan(plan.Name, false))
		}
	}
})
