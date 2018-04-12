.PHONY: lifecycle_tests

###### Help ###################################################################

help:
	@echo '    deps ...................................... installs dependencies'
	@echo '    go-vet .................................... runs go vet in source code'
	@echo '    test ...................................... runs test'

###### Dependencies ###########################################################

deps:
	git submodule update --init --recursive
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

###### Tools ###############################################################

go-vet:
	go vet `go list ./... | grep -v vendor`

test:
	ginkgo --trace -randomizeSuites=true -randomizeAllSpecs=true -keepGoing=true -failOnPending tests
