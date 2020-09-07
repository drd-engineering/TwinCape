package environments_test

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/drd-engineering/TwinCape/environments"
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	environments.Set("RELEASE_TYPE", "localhost")
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	return func(t *testing.T) {
		os.Clearenv()
	}
}
func TestLoadEnvironmentVariableFile(t *testing.T) {
	t.Run(
		"success load", func(t *testing.T) {
			set := setupTestCase(t)
			defer set(t)
			environments.LoadEnvironmentVariableFile()
		},
	)
	t.Run(
		"failed load", func(t *testing.T) {
			environments.LoadEnvironmentVariableFile()
		},
	)
}

func TestGet(t *testing.T) {
	set := setupTestCase(t)
	defer set(t)
	environments.LoadEnvironmentVariableFile()
	item := environments.Get("RELEASE_TYPE")
	assert.Equal(t, "localhost", item, "should return \"localhost\" as the result of get environment")
}

func TestSet(t *testing.T) {
	set := setupTestCase(t)
	defer set(t)
	environments.LoadEnvironmentVariableFile()
	environments.Set("RELEASE_TYPE", "test")
	item := environments.Get("RELEASE_TYPE")
	assert.Equal(t, "test", item, "should return \"test\" as the result of get environment")
}
