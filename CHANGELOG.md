# upcoming release (TBD)

Backwards-incompatible changes:

- `holo-test` has been removed from the public interface. Plugins that wish to use it are advised to vendor it from this
  repo into their own.

Changes:

- `holo`, `holo-files`, `holo-ssh-keys` and `holo-users-groups` have been merged into a single binary, thus massively
  reducing total installation size.
- Fix a bug in `holo-test` where tests could fail because of randomized names of temporary directories.
- Install the ALPM hook in the standard location.
- When Holo is installed via `go get`, show the version string "unknown" instead of an empty string.

Packagers beware:

- Add `Provides` and `Replaces` package relations from this package to `holo-ssh-keys` and `holo-users-groups`.

Notes on commit history:

- The history of the `holo-ssh-keys` repo can be found below the second parent of the merge commit e1e3d2e3d3826ddb2971f1e78b20b4dd467f3e28.
  The following tags were in that repo at the time of merging:

    a62c7288c464cbe359d9d0a17bc4c9b8556e3461 v1.0
    b4a2dc668e0a5caa3d80dc5b628acb814926330f v1.1
    453ac3fb6699e877584ba6e3027a6c88d206765f v1.2
    e6e25242795166cbcb18eedbbc7bb58250122027 v1.2.1

- The history of the `holo-users-groups` repo can be found below the second parent of the merge commit 2d6e87e41d62abced6f5c08428e7bea523cfb5a4.
  The following tags were in that repo at the time of merging:

    aced8d55b4dae5ef84eb3a4fde87240176498638 v1.0-beta.1
    c843bb2a66160d5c6371ec11af968985c6742d33 v1.0
    881c382776c48e543885662da74ea70a4f878793 v1.1
    be71e6b0415a3071f415429b597297c8bb3ab153 v1.2
    95f3792d73c7a82c923a43ebb81e537bde973618 v1.3
    66f46701aafbaa65397c2e4d68f8773f70272018 v2.0
    722e3be714ce5371e808e21d4af97c7dd8cdf9bd v2.0.1
    88f9154e0034ffff01fc686bcfd6ea970a676f2a v2.1
    f19426859936d15e43e6da5a3b28f11acd5f95e8 v2.1.1

# v1.3.1 (2017-03-19)

Bugfixes:

- Fix a bug where, on Arch Linux, the post-installation hook could get confused when Holo sorted entity names differently than sort(1) did.

# v1.3 (2017-03-18)

Special thanks to new contributor @LukeShu who did a lot of the hard work that went into this release, both in terms of
new features, boring refactoring work and documentation proof-reading.

Changes:

- `holo-files` now allows for fast-forwarding: When the computed content of a target file changes, but that change has
  already been done by the user, `holo-files` will now skip writing the target file and just update
  `/var/lib/holo/files/provisioned` instead of complaining that the target file does not match the previously
  provisioned content. (#24)
- When invoking Holo, plugin IDs can be used as selectors. For example, `holo apply ssh-keys` will apply all entities
  from the `holo-ssh-keys` plugin.

Bugfixes:

- Bring the scrubbing logic in line with the applying logic:
  - When a resource file is deleted while the target base is updated, restore the updated target base instead of the old
    one. (#16)
  - When a resource file is deleted and the saved version (`.pacsave`, `.rpmsave`, `.dpkg-old`) has been changed by the
    user, do not delete it. (#29)
  - Scrubbing has become more resilient against filesystem errors. When some file cannot be cleaned up, it will report
    that and keep going as much as possible. This is useful because Holo will forget about the entity once it is
    scrubbed, so the user should be informed about which actions remain to properly clean up the target file.
  - On Arch Linux, `.pacsave.N` files are now handled properly, similar to the existing handling for `.pacsave` files.
- Make sure that the cache directory (usually at `/tmp/holo.$$/`) is cleaned up even when an operation fails. (#20)
- Fix `make clean` to run correctly when the source is extracted from a tarball rather than cloned from git.
- Various fixes to `make check` to avoid false negatives.

Miscellaneous:

- Various internal refactorings.
- The documentation was proof-read and clarified in various locations.
- The test suite now checks code coverage.
- There are some files in `debian/` which should make it pretty easy to make a Debian package for Holo if anyone is
  interested in submitting it to Debian, Ubuntu etc.
- Releases are now signed by GPG key `0xD6019A3E17CA2D96`.

# v1.2.1 (2016-05-25)

Bugfixes:

- Fix installed holorc snippet.

# v1.2 (2016-05-25)

New features:

- Support modularized configuration in `/etc/holorc.d/*`, mainly to simplify the installation process for plugins. (#15)
- Add `--porcelain` option to `holo scan`.
- On Arch Linux with pacman >= 5.0, install a post-installation hook to handle .pacnew files automatically.

Miscellaneous:

- Strip binaries during build. (#14)

# v1.1.1 (2016-04-11)

Changes:

- Don't acquire the `/run/holo.pid` lockfile for readonly operations. This esp. unbreaks non-privileged usage of Holo e.g. for shell autocompletions because `/run` is writable only by root. (#13)
- Missing plugins are not a fatal error anymore. This fixes a logic deadlock during plugin uninstallation.

# v1.1 (2016-04-09)

Backward-incompatible changes:

- Packagers beware: The plugin interface version increases from **2** to **3**. (#12)
- Plugin developers: Update your plugins to
  - understand the new `info` command,
  - use the new semantics of the `diff` command (which now shall report files for diffing, instead of computing the diff itself),
  - and replace custom error messages with the new `requires --force to (update|restore)` messages where appropriate.

New features:

- Most output is now colorized appropriately, especially diffs from `holo diff`.
- `holo apply` without `--force` will now show a diff when the entity has manual changes that only `holo apply --force` will overwrite.
- Add a lockfile (`/run/holo.pid`) to prevent multiple parallel runs. (#9)
- `holo-test` now generates `colored-*-output` artifacts to allow plugin developers to inspect the colorized output of Holo. These artifacts are not validated against `expected-*-output`.

# v1.0.1 (2015-12-28)

Changes:

- Fix an edge case in `holo-files` which could cause source files to be applied in a different order than that reported by `holo scan`.
- Fix several glitches and inconsistences in the manpages. (Shout-out to @S1FeHa for proof-reading.)

# v1.0 (2015-12-18)

New features since Beta 2:

- Entities can now be identified by their source files. (#4)

Further changes since Beta 2:

- The name format for file entities has changed, from e.g. `/etc/foo.conf` to `file:/etc/foo.conf`.
- The manpages have been updated to describe the new plugin system.
- Fix a bug which caused unchanged target files to be reported during `holo apply --force`.

Plugin interface changes:

- The `HOLO_ROOT_DIR` variable is now always set, by default to `/`.
- The new `SOURCE:` directive can be used to link entities to their source files.
- The plugin interface version has increased to `HOLO_API_VERSION=2`.

# v1.0-beta.2 (2015-12-04)

Bugfixes:

- install `holorc` in the right path
- install `holo-test` to the right `$PATH`

# v1.0-beta.1 (2015-12-03)

Holo has been refactored into a plugin-based structure. The capabilities for [provisioning user accounts and groups][ug]
and [running custom provisioning scripts][rs] must now be installed separately.

**Backwards-incompatible changes:** A lot of filesystem paths change to follow the new plugin-based structure.

```
/usr/share/holo/{repo => files}/
/usr/share/holo/{provision => run-scripts}/
/usr/share/holo/{ => users-groups}/*.toml
/var/lib/holo/{ => files}/base/
/var/lib/holo/{ => files}/provisioned/
```

When updating, update all your configuration packages at the same time (to move stuff below `/usr/share/holo` into the
new locations), and take a backup of `/var/lib/holo` as the target bases will _definitely_ be messed up during the
update. Recipe:

```
cd /var/lib/holo
tar cf backup.tar base provisioned
[update Holo, install required plugins, update configuration packages for new paths]
cd /var/lib/holo/files
tar xf ../backup.tar
holo apply --force
cd /var/lib/holo
rm -r backup.tar base provisioned
```

Further changes:

- Optimize application algorithm: When the effect of the holoscript is overridden by a later repository entry that is a
  plain file, the holoscript is skipped entirely.

Known issues with this release:

- `make install` will put the holorc into the wrong place (`/etc/holo/holorc` instead of `/etc/holorc`).

This is the first release with the new split repository layout. Previous releases can be found [in the attic][ar].

[ug]: https://github.com/holocm/holo-users-groups
[rs]: https://github.com/holocm/holo-run-scripts
[ar]: https://github.com/holocm/holo-attic/releases
