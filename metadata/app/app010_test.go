package app

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp010_RenameEnvironment(t *testing.T) {
	cases := []struct {
		name           string
		from           string
		to             string
		shouldExist    []string
		shouldNotExist []string
	}{
		{
			name: "rename",
			from: "default",
			to:   "renamed",
			shouldExist: []string{
				"/environments/renamed/main.jsonnet",
			},
			shouldNotExist: []string{
				"/environments/default",
			},
		},
		{
			name: "rename to nested",
			from: "default",
			to:   "default/nested",
			shouldExist: []string{
				"/environments/default/nested/main.jsonnet",
			},
			shouldNotExist: []string{
				"/environments/default/main.jsonnet",
			},
		},
		{
			name: "un-nest",
			from: "us-east/test",
			to:   "us-east",
			shouldExist: []string{
				"/environments/us-east/main.jsonnet",
			},
			shouldNotExist: []string{
				"/environments/us-east/test",
			},
		},
	}

	for _, tc := range cases {
		withApp010Fs(t, "app010_app.yaml", func(fs afero.Fs) {
			t.Run(tc.name, func(t *testing.T) {
				app, err := NewApp010(fs, "/")
				require.NoError(t, err)

				err = app.RenameEnvironment(tc.from, tc.to)
				require.NoError(t, err)

				for _, p := range tc.shouldExist {
					checkExist(t, fs, p)
				}

				for _, p := range tc.shouldNotExist {
					checkNotExist(t, fs, p)
				}

				_, err = app.Environment(tc.from)
				assert.Error(t, err)

				_, err = app.Environment(tc.to)
				assert.NoError(t, err)
			})
		})
	}
}

func TestApp0101_Environments(t *testing.T) {
	withApp010Fs(t, "app010_app.yaml", func(fs afero.Fs) {
		app, err := NewApp010(fs, "/")
		require.NoError(t, err)

		expected := EnvironmentSpecs{
			"default": &EnvironmentSpec{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.7.0",
				Path:              "default",
			},
			"us-east/test": &EnvironmentSpec{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.7.0",
				Path:              "us-east/test",
			},
			"us-west/test": &EnvironmentSpec{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.7.0",
				Path:              "us-west/test",
			},
			"us-west/prod": &EnvironmentSpec{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.7.0",
				Path:              "us-west/prod",
			},
		}
		envs, err := app.Environments()
		require.NoError(t, err)

		require.Equal(t, expected, envs)
	})
}

func TestApp010_Environment(t *testing.T) {
	cases := []struct {
		name    string
		envName string
		isErr   bool
	}{
		{
			name:    "existing env",
			envName: "us-east/test",
		},
		{
			name:    "invalid env",
			envName: "missing",
			isErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withApp010Fs(t, "app010_app.yaml", func(fs afero.Fs) {
				app, err := NewApp010(fs, "/")
				require.NoError(t, err)

				spec, err := app.Environment(tc.envName)
				if tc.isErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					require.Equal(t, tc.envName, spec.Path)
				}
			})
		})
	}
}

func TestApp010_AddEnvironment(t *testing.T) {
	withApp010Fs(t, "app010_app.yaml", func(fs afero.Fs) {
		app, err := NewApp010(fs, "/")
		require.NoError(t, err)

		envs, err := app.Environments()
		require.NoError(t, err)

		envLen := len(envs)

		newEnv := &EnvironmentSpec{
			Destination: &EnvironmentDestinationSpec{
				Namespace: "some-namespace",
				Server:    "http://example.com",
			},
			Path: "us-west/qa",
		}

		k8sSpecFlag := "version:v1.8.7"
		err = app.AddEnvironment("us-west/qa", k8sSpecFlag, newEnv)
		require.NoError(t, err)

		envs, err = app.Environments()
		require.NoError(t, err)
		require.Len(t, envs, envLen+1)

		env, err := app.Environment("us-west/qa")
		require.NoError(t, err)
		require.Equal(t, "v1.8.7", env.KubernetesVersion)
	})
}

func TestApp010_AddEnvironment_empty_spec_flag(t *testing.T) {
	withApp010Fs(t, "app010_app.yaml", func(fs afero.Fs) {
		app, err := NewApp010(fs, "/")
		require.NoError(t, err)

		envs, err := app.Environments()
		require.NoError(t, err)

		envLen := len(envs)

		env, err := app.Environment("default")
		require.NoError(t, err)

		env.Destination.Namespace = "updated"

		err = app.AddEnvironment("default", "", env)
		require.NoError(t, err)

		envs, err = app.Environments()
		require.NoError(t, err)
		require.Len(t, envs, envLen)

		env, err = app.Environment("default")
		require.NoError(t, err)
		require.Equal(t, "v1.7.0", env.KubernetesVersion)
		require.Equal(t, "updated", env.Destination.Namespace)
	})
}

func TestApp010_RemoveEnvironment(t *testing.T) {
	withApp010Fs(t, "app010_app.yaml", func(fs afero.Fs) {
		app, err := NewApp010(fs, "/")
		require.NoError(t, err)

		_, err = app.Environment("default")
		require.NoError(t, err)

		err = app.RemoveEnvironment("default")
		require.NoError(t, err)

		app, err = NewApp010(fs, "/")
		require.NoError(t, err)

		_, err = app.Environment("default")
		require.Error(t, err)
	})
}

func withApp010Fs(t *testing.T, appName string, fn func(fs afero.Fs)) {
	ogLibUpdater := LibUpdater
	LibUpdater = func(fs afero.Fs, k8sSpecFlag string, libPath string, useVersionPath bool) (string, error) {
		return "v1.8.7", nil
	}

	defer func() {
		LibUpdater = ogLibUpdater
	}()

	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	fs := afero.NewBasePathFs(afero.NewOsFs(), dir)

	envDirs := []string{
		"default",
		"us-east/test",
		"us-west/test",
		"us-west/prod",
	}

	for _, dir := range envDirs {
		path := filepath.Join("/environments", dir)
		err := fs.MkdirAll(path, DefaultFolderPermissions)
		require.NoError(t, err)

		swaggerPath := filepath.Join(path, "main.jsonnet")
		stageFile(t, fs, "main.jsonnet", swaggerPath)
	}

	stageFile(t, fs, appName, "/app.yaml")

	fn(fs)
}
