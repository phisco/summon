package command

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cyberark/summon/secretsyml"
	. "github.com/smartystreets/goconvey/convey"
	_ "golang.org/x/net/context"
)

func TestConvertSubsToMap(t *testing.T) {
	Convey("Substitutions are returned as a map used later for interpolation", t, func() {
		input := []string{
			"policy=accounts-database",
			"environment=production",
		}

		expected := map[string]string{
			"policy":      "accounts-database",
			"environment": "production",
		}

		output := convertSubsToMap(input)

		So(output, ShouldResemble, expected)
	})
}

func TestFormatForEnvString(t *testing.T) {
	Convey("formatForEnv should return a KEY=VALUE string that can be appended to an environment", t, func() {
		Convey("For variables, VALUE should be the value of the secret", func() {
			spec := secretsyml.SecretSpec{
				Path: "mysql1/password",
				Tags: []secretsyml.YamlTag{secretsyml.Var},
			}
			envvar := formatForEnv(
				"dbpass",
				"mysecretvalue",
				spec,
				nil,
			)

			So(envvar, ShouldEqual, "dbpass=mysecretvalue")
		})
		Convey("For files, VALUE should be the path to a tempfile containing the secret", func() {
			temp_factory := NewTempFactory("")
			defer temp_factory.Cleanup()

			spec := secretsyml.SecretSpec{
				Path: "certs/webtier1/private-cert",
				Tags: []secretsyml.YamlTag{secretsyml.File},
			}
			envvar := formatForEnv(
				"SSL_CERT",
				"mysecretvalue",
				spec,
				&temp_factory,
			)

			s := strings.Split(envvar, "=")
			key, path := s[0], s[1]

			So(key, ShouldEqual, "SSL_CERT")

			// Temp path should exist
			_, err := os.Stat(path)
			So(err, ShouldBeNil)

			contents, _ := ioutil.ReadFile(path)

			So(string(contents), ShouldContainSubstring, "mysecretvalue")
		})
	})
}

func TestJoinEnv(t *testing.T) {
	Convey("adds a trailing newline", t, func() {
		result := joinEnv([]string{"foo", "bar"})
		So(result, ShouldEqual, "foo\nbar\n")
	})
}

func TestRunAction(t *testing.T) {
	Convey("Variable resolution correctly resolves variables", t, func() {
		expectedValue := "valueOfVariable"

		dir, err := ioutil.TempDir("", "summon")
		So(err, ShouldBeNil)
		if err != nil {
			return
		}
		defer os.RemoveAll(dir)

		tempFile := filepath.Join(dir, "outputFile.txt")

		err = runAction(&ActionConfig{
			Args:       []string{"bash", "-c", "echo -n \"$FOO\" > " + tempFile},
			YamlInline: "FOO: " + expectedValue,
		})

		code, err := returnStatusOfError(err)
		So(err, ShouldBeNil)
		So(code, ShouldEqual, 0)

		if err != nil || code != 0 {
			return
		}

		content, err := ioutil.ReadFile(tempFile)
		So(err, ShouldBeNil)
		if err != nil {
			return
		}

		So(string(content), ShouldEqual, expectedValue)
	})
}

func TestDefaultVariableResolution(t *testing.T) {
	Convey("Variable resolution correctly resolves variables", t, func() {
		expectedDefaultValue := "defaultValueOfVariable"

		dir, err := ioutil.TempDir("", "summon")
		So(err, ShouldBeNil)
		if err != nil {
			return
		}
		defer os.RemoveAll(dir)

		tempFile := filepath.Join(dir, "outputFile.txt")

		err = runAction(&ActionConfig{
			Args:       []string{"bash", "-c", "echo -n \"$FOO\" > " + tempFile},
			YamlInline: "FOO: !str:default='" + expectedDefaultValue + "'",
		})

		code, err := returnStatusOfError(err)
		So(err, ShouldBeNil)
		So(code, ShouldEqual, 0)

		if err != nil || code != 0 {
			return
		}

		content, err := ioutil.ReadFile(tempFile)
		So(err, ShouldBeNil)
		if err != nil {
			return
		}

		So(string(content), ShouldEqual, expectedDefaultValue)
	})
}

func TestDefaultVariableResolutionWithValue(t *testing.T) {
	Convey("Variable resolution correctly resolves variables", t, func() {
		expectedValue := "valueOfVariable"

		dir, err := ioutil.TempDir("", "summon")
		So(err, ShouldBeNil)
		if err != nil {
			return
		}
		defer os.RemoveAll(dir)

		tempFile := filepath.Join(dir, "outputFile.txt")

		err = runAction(&ActionConfig{
			Args:       []string{"bash", "-c", "echo -n \"$FOO\" > " + tempFile},
			YamlInline: "FOO: !str:default='something' " + expectedValue,
		})

		code, err := returnStatusOfError(err)
		So(err, ShouldBeNil)
		So(code, ShouldEqual, 0)

		if err != nil || code != 0 {
			return
		}

		content, err := ioutil.ReadFile(tempFile)
		So(err, ShouldBeNil)
		if err != nil {
			return
		}

		So(string(content), ShouldEqual, expectedValue)
	})
}

func TestReturnStatusOfError(t *testing.T) {
	Convey("returns no error as 0", t, func() {
		res, err := returnStatusOfError(nil)
		So(res, ShouldEqual, 0)
		So(err, ShouldBeNil)
	})

	Convey("returns ExitError as the wrapped exit status", t, func() {
		exit := exec.Command("false").Run()
		res, err := returnStatusOfError(exit)
		So(res, ShouldEqual, 1)
		So(err, ShouldBeNil)
	})

	Convey("returns other errors unchanged", t, func() {
		expected := errors.New("test")
		_, err := returnStatusOfError(expected)
		So(err, ShouldEqual, expected)
	})
}
