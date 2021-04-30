PKG_FILES=`go list ./... | sed -e 's=github.com/bitleak/go-lmstfy-client/=./='`

CCCOLOR="\033[37;1m"
MAKECOLOR="\033[32;1m"
ENDCOLOR="\033[0m"

.PHONY: all

setup:
	@bash scripts/setup.sh

teardown:
	@bash scripts/teardown.sh

test:
	@bash scripts/setup.sh
	@bash scripts/test.sh
	@bash scripts/teardown.sh

lint:
	@rm -rf lint.log
	@printf $(CCCOLOR)"Checking format...\n"$(ENDCOLOR)
	@go list ./... | sed -e 's=github.com/bitleak/go-lmstfy-client=.=' | xargs -n 1 gofmt -d -s 2>&1 | tee -a lint.log
	@[ ! -s lint.log ]
	@printf $(CCCOLOR)"Checking vet...\n"$(ENDCOLOR)
	@go list ./... | sed -e 's=github.com/bitleak/go-lmstfy-client=.=' | xargs -n 1 go vet
	@printf $(CCCOLOR)"GolangCI Lint...\n"$(ENDCOLOR)
	@golangci-lint run
