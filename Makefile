# application name
APP ?= orchid
# sanitizing app variable to become a valid go module name
MODULE = $(subst -,,$(APP))

RUN_ARGS ?= serve
COMMON_FLAGS ?= -v -mod=vendor

TEST_TIMEOUT ?= 3m
TEST_FLAGS ?= -failfast -timeout=$(TEST_TIMEOUT)
TEST_EXTRA_FLAGS ?=

OUTPUT_DIR ?= build

CODECOV_TOKEN ?=
COVERAGE_DIR ?= $(OUTPUT_DIR)/coverage

# used in `codecov.sh` script
export OUTPUT_DIR
export COVERAGE_DIR
export CODECOV_TOKEN

default: build

# initialize Go modules vendor directory
.PHONY: vendor
vendor:
	@go mod vendor

# clean up build directory
clean:
	@rm -rf $(OUTPUT_DIR)

# create build and coverage directories
prepare:
	@mkdir -p $(COVERAGE_DIR) > /dev/null 2>&1 || true

.PHONY: $(OUTPUT_DIR)/$(APP)
$(OUTPUT_DIR)/$(APP):
	go build $(COMMON_FLAGS) -o="$(OUTPUT_DIR)/$(APP)" cmd/$(MODULE)/*

# build application command-line
build: prepare vendor $(OUTPUT_DIR)/$(APP)

# execute "go run" against cmd
run:
	go run $(COMMON_FLAGS) cmd/$(MODULE)/* $(RUN_ARGS)

# running all test targets
test: test-unit test-e2e

# run unit tests
test-unit: prepare
	go test $(COMMON_FLAGS) $(TEST_FLAGS) $(TEST_EXTRA_FLAGS) \
		-coverprofile=$(COVERAGE_DIR)/coverage-unit.txt ./...

# run end-to-end tests
test-e2e:
	echo "TODO: include end-to-end tests here!"

# codecov.io test coverage report
codecov:
	./hack/codecov.sh
