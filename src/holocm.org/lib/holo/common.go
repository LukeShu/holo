package holo

type Runtime struct {
	APIVersion int

	RootDirPath string

	ResourceDirPath string
	StateDirPath    string
	CacheDirPath    string
}

type ApplyResult interface {
	isApplyResult()
	Exit()
}

var (
	// Indicates that the entity has been successfully modified to
	// be in the desired state.
	ApplyApplied ApplyResult = applyResult(applyApplied)

	// Indicates that there was an error modifying the entity.
	ApplyErr func(n int) ApplyResult = applyErr

	// Indicates that the entity is already in the desired state,
	// so no changes have been made. Holo will format its output
	// accordingly (at the time of this writing, by omitting the
	// entity from the output).
	ApplyAlreadyApplied ApplyResult = applyResult(applyAlreadyApplied)

	// Indicates that the entity was provisioned by this plugin,
	// but has been changed by a user or external application
	// since then.  Holo will output an error message indicating
	// that "--force" is needed to overwrite these manual changes.
	ApplyExternallyChanged ApplyResult = applyResult(applyExternallyChanged)

	// Indicateq that the entity was provisioned by this plugin,
	// but has been deleted by the user or external application
	// since then.  Holo will output an error message indicating
	// that "--force" is needed to overwrite these manual changes.
	ApplyExternallyDeleted ApplyResult = applyResult(applyExternallyDeleted)
)

type KV struct {
	Key, Val string
}

type Entity interface {
	EntityID() string
	EntitySource() []string
	EntityAction() string
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
	HoloApply(entityID string, force bool) ApplyResult

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
