default: bin/holo bin/holo-files
default: man/holorc.5 man/holo-plugin-interface.7 man/holo-test.7 man/holo.8 man/holo-files.8
.PHONY: default

GO_BUILDFLAGS :=
GO_LDFLAGS    := -s -w

src/holocm.org/cmd/holo/version.go: FORCE
	printf 'package main\n\nconst version = "%s"\n' "$$( ./util/find_version.sh)" | util/write-ifchanged $@

%/holo %/holo-files: FORCE src/holocm.org/cmd/holo/version.go
	GOPATH=$(dir $(abspath $*)) go install $(GO_BUILDFLAGS) --ldflags '$(GO_LDFLAGS)' holocm.org/cmd/holo holocm.org/cmd/holo-files

man:
	@mkdir $@

# manpages are generated using pod2man (which comes with Perl and therefore
# should be readily available on almost every Unix system)
man/%: doc/%.pod | man
	pod2man --name="$(shell echo $* | cut -d. -f1)" --section=$(shell echo $* | cut -d. -f2) \
		--center="Configuration Management" --release="Holo $(VERSION)" \
		$< $@

test: check # just a synonym
check: default
	@GOPATH=$(abspath .) go test holocm.org/cmd/holo/impl
	@env HOLO_BINARY=../../bin/holo bash util/holo-test holo $(sort $(wildcard test/??-*))
.PHONY: test check

install: default conf/holorc conf/holorc.holo-files util/holo-test util/autocomplete.bash util/autocomplete.zsh
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/base"
	install -d -m 0755 "$(DESTDIR)/var/lib/holo/files/provisioned"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo"
	install -d -m 0755 "$(DESTDIR)/usr/share/holo/files"
	install -D -m 0644 conf/holorc            "$(DESTDIR)/etc/holorc"
	install -D -m 0644 conf/holorc.holo-files "$(DESTDIR)/etc/holorc.d/10-files"
	install -D -m 0755 bin/holo               "$(DESTDIR)/usr/bin/holo"
	install -D -m 0755 bin/holo-files         "$(DESTDIR)/usr/lib/holo/holo-files"
	install -D -m 0755 util/holo-test         "$(DESTDIR)/usr/bin/holo-test"
	install -D -m 0644 util/autocomplete.bash "$(DESTDIR)/usr/share/bash-completion/completions/holo"
	install -D -m 0644 util/autocomplete.zsh  "$(DESTDIR)/usr/share/zsh/site-functions/_holo"
	install -D -m 0644 man/holorc.5                "$(DESTDIR)/usr/share/man/man5/holorc.5"
	install -D -m 0644 man/holo.8                  "$(DESTDIR)/usr/share/man/man8/holo.8"
	install -D -m 0644 man/holo-files.8            "$(DESTDIR)/usr/share/man/man8/holo-files.8"
	install -D -m 0644 man/holo-test.7             "$(DESTDIR)/usr/share/man/man7/holo-test.7"
	install -D -m 0644 man/holo-plugin-interface.7 "$(DESTDIR)/usr/share/man/man7/holo-plugin-interface.7"
	env DESTDIR=$(DESTDIR) ./src/distribution-integration/install.sh
.PHONY: install

clean:
	rm -fr -- bin/holo bin/holo-files man test/*/{target,tree} test/*/{colored-,}{apply,apply-force,diff,scan}-output
	rm -fr -- bin/holo bin/holo-files man
	rm -fr -- test/*/target test/*/tree test/*/[acds]*-output
.PHONY: clean

.PHONY: FORCE
