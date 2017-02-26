pkg = github.com/holocm/holo
bins = holo holo-files tinyholo
mans = holorc.5 holo-plugin-interface.7 holo-test.7 holo.8 holo-files.8

GO            := GOPATH=$(CURDIR)/.go-workspace GOBIN=$(CURDIR)/build go
GO_BUILDFLAGS :=
GO_LDFLAGS    := -s -w
GO_TESTFLAGS  := -covermode=count
GO_DEPS       := $(GO) list -f '{{.ImportPath}}{{"\n"}}{{join .Deps "\n"}}'

all: $(addprefix build/,$(bins))
all: $(addprefix build/man/,$(mans))
.PHONY: all

.version: FORCE
	./util/find_version.sh | util/write-ifchanged $@

cmd/holo/version.go: .version
	printf 'package main\n\nfunc init() {\n\tversion = "%s"\n}\n' "$$(cat $<)" > $@

$(addprefix %/,$(bins)): FORCE cmd/holo/version.go
	$(GO) install $(GO_BUILDFLAGS) --ldflags '$(GO_LDFLAGS)' $(addprefix $(pkg)/cmd/,$(bins))
build/%.test: build/% cmd/%/main_test.go
	$(GO) test -c -o $@ $(GO_TESTFLAGS) -coverpkg $$($(GO_DEPS) $(pkg)/cmd/$*|grep ^$(pkg)|tr '\n' ,|sed 's/,$$//') $(pkg)/cmd/$*

build/man:
	mkdir -p $@

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
build/man/%: doc/%.pod .version | build/man
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) \
		--center="Configuration Management" --release="Holo $$(cat .version)" \
		$< $@

test: check # just a synonym
check: test/cov.html test/cov.func.txt
.PHONY: test check
test/cov.cov: clean-tests $(foreach b,$(bins),build/$b.test)
	rm -f -- test/cov/* test/cov.*
	@if s="$$(gofmt -l cmd 2>/dev/null)"                        && test -n "$$s"; then printf ' => %s\n%s\n' gofmt  "$$s"; false; fi
	@if s="$$(find cmd -type d -exec golint {} \; 2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' golint "$$s"; false; fi
	$(GO) test $(GO_TESTFLAGS) -coverprofile=test/cov/holo-output.cov $(pkg)/cmd/holo/output
	HOLO_BINARY=../../build/holo.test     HOLO_TEST_COVERDIR=$(abspath test/cov) util/holo-test holo $(sort $(wildcard test/??-*))
	HOLO_BINARY=../../build/tinyholo.test HOLO_TEST_COVERDIR=$(abspath test/cov) util/holo-test holo $(sort $(wildcard test/??-*))
	util/gocovcat.go test/cov/*.cov > test/cov.cov
%.html: %.cov
	$(GO) tool cover -html $< -o $@
%.func.txt: %.cov
	$(GO) tool cover -func $< -o $@

install-holo: all conf/holorc util/holo-test util/autocomplete.bash util/autocomplete.zsh
	install -d -m 0755 "$(DESTDIR)/usr/share/holo"
	install -D -m 0755 build/holo             "$(DESTDIR)/usr/bin/holo"
	install -D -m 0644 build/man/holorc.5                "$(DESTDIR)/usr/share/man/man5/holorc.5"
	install -D -m 0644 build/man/holo.8                  "$(DESTDIR)/usr/share/man/man8/holo.8"
	install -D -m 0644 build/man/holo-plugin-interface.7 "$(DESTDIR)/usr/share/man/man7/holo-plugin-interface.7"
	install -D -m 0644 conf/holorc            "$(DESTDIR)/etc/holorc"
	install -D -m 0644 util/autocomplete.bash "$(DESTDIR)/usr/share/bash-completion/completions/holo"
	install -D -m 0644 util/autocomplete.zsh  "$(DESTDIR)/usr/share/zsh/site-functions/_holo"
install-holo-test: all util/holo-test
	install -D -m 0755 util/holo-test         "$(DESTDIR)/usr/bin/holo-test"
	install -D -m 0644 build/man/holo-test.7  "$(DESTDIR)/usr/share/man/man7/holo-test.7"
install-holo-files: all conf/holorc.holo-files
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/files"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/base"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/provisioned"
	install -D -m 0644 conf/holorc.holo-files "$(DESTDIR)/etc/holorc.d/10-files"
	install -D -m 0755 build/holo-files       "$(DESTDIR)/usr/lib/holo/holo-files"
	install -D -m 0644 build/man/holo-files.8 "$(DESTDIR)/usr/share/man/man8/holo-files.8"
install: install-holo install-holo-test install-holo-files
	DESTDIR=$(DESTDIR) util/distribution-integration/install.sh
.PHONY: install install-holo install-holo-test install-holo-files

clean: clean-tests
	rm -fr -- build/ .go-workspace/pkg/
	rm -f -- .version cmd/holo/version.go
clean-tests:
	rm -fr -- test/*/{target,tree,{colored-,}{apply,apply-force,diff,scan}-output}
	rm -f -- test/cov.* test/cov/*
.PHONY: clean clean-tests

.PHONY: FORCE
.DELETE_ON_ERROR:
