#!/bin/bash
set -e

source /var/vcap/packages/golang-1.10-linux/bosh/runtime.env

. /var/vcap/jobs/smoke-tests/bin/change-permissions
. /var/vcap/jobs/smoke-tests/bin/permissions-test

export GOPATH=/var/vcap/packages/cf-rabbitmq-smoke-tests
export PATH=/var/vcap/packages/cf-cli-6-linux/bin:$GOPATH/bin:$GOROOT/bin:$PATH
export REPO_NAME=github.com/pivotal-cf/cf-rabbitmq-smoke-tests
export REPO_DIR=${GOPATH}/src/${REPO_NAME}

export CONFIG_PATH=/var/vcap/jobs/smoke-tests/config.json

export CF_DIAL_TIMEOUT=11
export SMOKE_TESTS_TIMEOUT=1h

pushd ${REPO_DIR}
  echo "Running multitenant smoke tests"
  go install -v github.com/onsi/ginkgo/ginkgo
  ginkgo -v --trace -randomizeSuites=true -randomizeAllSpecs=true -keepGoing=true ---timeout="$SMOKE_TESTS_TIMEOUT" -failOnPending tests
popd
