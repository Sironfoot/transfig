package transfig_test

import (
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/sironfoot/transfig"
)

func copyFile(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

func TestLoad_NonPointer(t *testing.T) {
	var configuration complex

	err := transfig.Load("complex.json", "", configuration)

	// assert
	if err != transfig.ErrConfigDataNotPointer {
		t.Errorf("should have returned error: %s", transfig.ErrConfigDataNotPointer)
	}
}

func TestLoad_NoEnvironment(t *testing.T) {
	// arrange
	var actualConfig complex

	// act
	err := transfig.Load("complex.json", "dev", &actualConfig)

	// assert
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expectedConfig, actualConfig) {
		t.Errorf("expected and actual config are different.\nExpected:\n%v\n\nActual:\n%v", expectedConfig, actualConfig)
	}
}

func TestLoad_PrimaryFileNotExist(t *testing.T) {
	// arrange
	var actualConfig complex

	// act
	err := transfig.Load("notExsts.json", "dev", &actualConfig)

	// assert
	if err != transfig.ErrPrimaryConfigFileNotExist {
		t.Errorf("err expected: \"%s\" but got: \"%s\"", transfig.ErrPrimaryConfigFileNotExist, err)
	}
}
