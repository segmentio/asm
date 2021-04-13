dstdir := $(CURDIR)
srcdir := $(CURDIR)/build

sources := \
	$(wildcard $(srcdir)/*_asm.go \
	$(srcdir)/bloom/*_asm.go)

targets := \
	$(patsubst $(srcdir)/%_asm.go,$(dstdir)/%_amd64.s,$(sources)) \
	$(patsubst $(srcdir)/%_asm.go,$(dstdir)/%_amd64.go,$(sources))

build: $(targets)

$(dstdir)/%_amd64.s $(dstdir)/%_amd64.go: $(srcdir)/%_asm.go
	cd build && go run $(patsubst $(CURDIR)/build/%,%,$<) \
		-pkg   $(notdir $(realpath $(dir $<))) \
		-out   ../$(patsubst $(CURDIR)/%,%,$(patsubst $(srcdir)/%_asm.go,$(dstdir)/%_amd64.s,$<)) \
		-stubs ../$(patsubst $(CURDIR)/%,%,$(patsubst $(srcdir)/%_asm.go,$(dstdir)/%_amd64.go,$<))

.PHONY: build
