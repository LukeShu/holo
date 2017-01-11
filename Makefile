bins = holo holo-files
mans = holorc.5 holo-plugin-interface.7 holo-test.7 holo.8 holo-files.8

GO_BUILDFLAGS :=
GO_LDFLAGS    := -s -w

all: $(addprefix bin/,$(bins))
all: $(addprefix man/,$(mans))
.PHONY: all

src/holocm.org/cmd/holo/version.go: FORCE
	printf 'package main\n\nconst version = "%s"\n' "$$( ./util/find_version.sh)" | util/write-ifchanged $@

%/holo %/holo-files: FORCE src/holocm.org/cmd/holo/version.go
	GOPATH=$(dir $(abspath $*)) go install $(GO_BUILDFLAGS) --ldflags '$(GO_LDFLAGS)' holocm.org/cmd/holo holocm.org/cmd/holo-files

man:
	mkdir $@

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
man/%: doc/%.pod | man
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) \
		--center="Configuration Management" --release="Holo $(VERSION)" \
		$< $@
test: check # just a synonym
check: all util/holo-test
	GOPATH=$(abspath .) go test holocm.org/cmd/holo/impl
	HOLO_BINARY=../../bin/holo bash util/holo-test holo $(sort $(wildcard test/??-*))
.PHONY: test check

install-holo: all conf/holorc src/holo-test util/autocomplete.bash util/autocomplete.zsh
	install -d -m 0755 "$(DESTDIR)/usr/share/holo"
	install -D -m 0755 bin/holo               "$(DESTDIR)/usr/bin/holo"
	install -D -m 0644 man/holorc.5                "$(DESTDIR)/usr/share/man/man5/holorc.5"
	install -D -m 0644 man/holo.8                  "$(DESTDIR)/usr/share/man/man8/holo.8"
	install -D -m 0644 man/holo-plugin-interface.7 "$(DESTDIR)/usr/share/man/man7/holo-plugin-interface.7"
	install -D -m 0644 conf/holorc            "$(DESTDIR)/etc/holorc"
	install -D -m 0644 util/autocomplete.bash "$(DESTDIR)/usr/share/bash-completion/completions/holo"
	install -D -m 0644 util/autocomplete.zsh  "$(DESTDIR)/usr/share/zsh/site-functions/_holo"
install-holo-test: all util/holo-test
	install -D -m 0755 util/holo-test   "$(DESTDIR)/usr/bin/holo-test"
	install -D -m 0644 man/holo-test.7  "$(DESTDIR)/usr/share/man/man7/holo-test.7"
install-holo-files: all conf/holorc.holo-files
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/files"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/base"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/provisioned"
	install -D -m 0644 conf/holorc.holo-files "$(DESTDIR)/etc/holorc.d/10-files"
	install -D -m 0755 bin/holo-files         "$(DESTDIR)/usr/lib/holo/holo-files"
	install -D -m 0644 man/holo-files.8       "$(DESTDIR)/usr/share/man/man8/holo-files.8"
install: install-holo install-holo-test instal-holo-files
	DESTDIR=$(DESTDIR) distribution-integration/install.sh
.PHONY: install install-holo install-holo-test instal-holo-files

clean:
	rm -fr -- pkg bin man
	rm -fr -- test/*/{target,tree,{colored-,}{apply,apply-force,diff,scan}-output}
.PHONY: clean

.PHONY: FORCE
.DELETE_ON_ERROR:
