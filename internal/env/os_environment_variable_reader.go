package env

import "os"

type OSEnvironmentVariableReader struct {
}

func NewOSEnvironmentVariableReader() *OSEnvironmentVariableReader {
	return &OSEnvironmentVariableReader{}
}

func (kcr *OSEnvironmentVariableReader) Read(name string) (string, bool) {
	return os.LookupEnv(name)
}
