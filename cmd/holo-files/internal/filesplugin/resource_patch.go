package filesplugin
	"github.com/holocm/holo/cmd/holo-files/internal/fileutil"
func (resource Patchfile) ApplyTo(entityBuffer fileutil.FileBuffer) (fileutil.FileBuffer, error) {
	//  1. since fileutil.FileBuffer.Write removes the file and then
		return fileutil.FileBuffer{}, err
		return fileutil.FileBuffer{}, err
		return fileutil.FileBuffer{}, err
		return fileutil.FileBuffer{}, fmt.Errorf("execution failed: %s: %s", strings.Join(cmd.Args, " "), err.Error())
	targetBuffer, err := fileutil.NewFileBuffer(targetPath)
		return fileutil.FileBuffer{}, err