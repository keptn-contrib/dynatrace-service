package env

import "os"

type OSEnvironmentVariableReader struct {
}

func NewOSEnvironmentVariableReader() (*OSEnvironmentVariableReader, error) {
	osEnvironmentVariableReader := &OSEnvironmentVariableReader{}

	return osEnvironmentVariableReader, nil
}

func (kcr *OSEnvironmentVariableReader) Read(name string) (string, bool) {
	return os.LookupEnv(name)
}
