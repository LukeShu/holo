package holo

import (
	"io"
)

type Runtime struct {
	APIVersion int

	RootDirPath string

	ResourceDirPath string
	StateDirPath    string
	CacheDirPath    string
}

type ApplyResult interface {
	isApplyResult()
	ExitCode() int
}

type ApplyMessage struct {
	msg string
}

type ApplyError interface {
	ApplyResult
	error
}

var (
	// Indicates that the entity has been successfully modified to
	// be in the desired state.
	ApplyApplied = applyApplied{}

	// Indicates that the entity is already in the desired state,
	// so no changes have been made. Holo will format its output
	// accordingly (at the time of this writing, by omitting the
	// entity from the output).
	ApplyAlreadyApplied = ApplyMessage{"not changed\n"}

	// Indicates that the entity was provisioned by this plugin,
	// but has been changed by a user or external application
	// since then.  Holo will output an error message indicating
	// that "--force" is needed to overwrite these manual changes.
	ApplyExternallyChanged = ApplyMessage{"requires --force to overwrite\n"}

	// Indicateq that the entity was provisioned by this plugin,
	// but has been deleted by the user or external application
	// since then.  Holo will output an error message indicating
	// that "--force" is needed to overwrite these manual changes.
	ApplyExternallyDeleted = ApplyMessage{"requires --force to restore\n"}
)

type KV struct {
	Key, Val string
}

type Entity interface {
	EntityID() string
	EntitySource() []string
	EntityAction() (verb, reason string)
	EntityUserInfo() []KV
}

type Plugin interface {
	// Return metadata about the plugin itself.
	//
	// "MIN_API_VERSION" and "MAX_API_VERSION"
	HoloInfo() map[string]string

	// Scan Runtime.ResourceDirPath and return a list of entities
	// that this plugin can provision.
	HoloScan() ([]Entity, error)

	// Provision entityID
	HoloApply(entityID string, force bool, stdout, stderr io.Writer) ApplyResult

	// Return two file paths.  The file pointed to by the first is
	// a representation of entity in the desired provisioned
	// state; the second is a representation of the entity in its
	// current state.
	//
	// If either state is "not existing", an empty string may be
	// returned for one state.  If the entity does not have a
	// meaningful textual representation, then two empty strings
	// should be returned.
	HoloDiff(entityID string) (string, string)
}
