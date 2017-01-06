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
	HoloScan(stderr io.Writer) ([]Entity, error)

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
	HoloDiff(entityID string, stderr io.Writer) (string, string)
}
