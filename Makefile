
.PHONY: lint vet fmt fmtcheck pretest test coverage build cross-build clean savedeps

SRCS = $(shell find . -name '*.go' | grep -v /vendor/)
PKGS = $(shell glide nv)
BUILD_DIR = _build
TEST_REPORT = $(BUILD_DIR)/test_report.xml
COVERAGE_REPORT_HTML = $(BUILD_DIR)/coverage_report.html
COVERAGE_REPORT_XML = $(BUILD_DIR)/coverage_report.xml

BASE_OUTPUT_NAME = $(shell basename $(CURDIR))
TEST_TMP = test_tmp
COVERAGE_TMP = $(BUILD_DIR)/coverage_tmp.json

all: test build

mk_build_dir:
		@ mkdir -p $(BUILD_DIR)

lint:
		@ $(foreach file,$(PKGS),golint $(file) || exit;)

vet:
		@ $(foreach pkg,$(PKGS),go vet $(pkg) || exit;)

fmt:
		@ gofmt -s -w $(SRCS)

fmtcheck:
		@ gofmt -s -d $(SRCS)

pretest: vet fmtcheck mk_build_dir

test: pretest
		@ go test -v $(PKGS) > $(TEST_TMP); \
					exit_status=$$?; \
							cat $(TEST_TMP); \
									cat $(TEST_TMP) | go-junit-report > $(TEST_REPORT); \
											rm $(TEST_TMP); \
													exit $$exit_status

coverage: pretest
		@ gocov test $(PKGS) > $(COVERAGE_TMP)
			@ gocov report $(COVERAGE_TMP)
				@ gocov-html $(COVERAGE_TMP) > $(COVERAGE_REPORT_HTML)
					@ cat $(COVERAGE_TMP) | gocov-xml > $(COVERAGE_REPORT_XML)
						@ rm -f $(COVERAGE_TMP)

build: mk_build_dir
		@ go build
			@ find . -maxdepth 1 -name "$(BASE_OUTPUT_NAME)*" -and -not -name '*.go' -and -not -name "*.iml" -exec mv {} $(BUILD_DIR) \;

cross-build: mk_build_dir
		@ gox -osarch="darwin/amd64" -osarch="linux/amd64" -osarch="windows/amd64" $(PKGS)
			@ find . -maxdepth 1 -name "$(BASE_OUTPUT_NAME)*" -and -not -name '*.go' -and -not -name "*.iml" -exec mv {} $(BUILD_DIR) \;

savedeps:
		@ godep save -t ./...

clean:
		@ go clean
			@ rm -Rf $(BUILD_DIR)
