	"io"
func (resource Patchfile) ApplyTo(entityBuffer fileutil.FileBuffer, stdout, stderr io.Writer) (fileutil.FileBuffer, error) {
	cmd.Stdout = stderr
	cmd.Stderr = stderr