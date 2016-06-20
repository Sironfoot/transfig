package transfig_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/sironfoot/transfig"
)

func TestWithComments(t *testing.T) {
	err := copyFile("withComments.test.json", "withComments.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.Remove("withComments.test.json")
		if err != nil {
			t.Fatal(err)
		}
	}()

	var config complex
	err = transfig.Load("withComments.json", "test", &config)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expectedConfig, config) {
		t.Errorf("expected and actual config are different.\nExpected:\n%v\n\nActual:\n%v", expectedConfig, config)
	}
}
