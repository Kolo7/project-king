
.DEFAULT_GOAL := all

.PHONY: all
all: tidy build

ROOT_PACKAGE=.


define USAGE_OPTIONS

Options:

endef
export USAGE_OPTIONS

# 临时的将零散的变量定义在这
include scripts/make-rules/common.mk # make sure include common.mk at the first include line
include scripts/make-rules/golang.mk
###########

## tidy: golang整理依赖库
.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: build
build: go.build



# help将会自动解析所有伪target带有##开头的注释，生成帮助文档
## help: Show this help info.
.PHONY: help
help: Makefile
	@echo -e "\nUsage: make <TARGETS> <OPTIONS> ...\n\nTargets:"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
	@echo "$$USAGE_OPTIONS"