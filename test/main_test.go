package transfig_test

import (
	"reflect"
	"testing"

	"github.com/sironfoot/transfig"
)

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
