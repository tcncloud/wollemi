package please

type Builder interface {
	Ctl
	Parse(string, []byte) (File, error)
	NewFile(string) File
	NewRule(string, string) Rule
	Write(File) error
}
