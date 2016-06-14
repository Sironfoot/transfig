package transfig_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/sironfoot/transfig"
)

type varLengthSlices struct {
	Strings []string  `json:"strings"`
	Floats  []float64 `json:"floats"`
	Ints    []int     `json:"ints"`
	Bools   []bool    `json:"bools"`
}

func TestLoad_VariableLengthSlices(t *testing.T) {
	// arrange
	var actualConfig varLengthSlices

	altConfigString := `
    {
        "strings": [],
        "floats": [ 1.2 ],
        "ints": [ 1, 3, 5, 7, 9 ],
        "bools": [ true, false, false, false, true, false, true, true ]
    }`

	var expectedAltConfig varLengthSlices
	err := json.Unmarshal([]byte(altConfigString), &expectedAltConfig)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile("varLengthSlices.test.json", []byte(altConfigString), 0644)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = os.Remove("varLengthSlices.test.json")
		if err != nil {
			t.Fatal(err)
		}
	}()

	// act
	err = transfig.Load("varLengthSlices.json", "test", &actualConfig)

	// assert
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expectedAltConfig, actualConfig) {
		t.Errorf("expected and actual config are different.\nExpected:\n%v\n\nActual:\n%v", expectedAltConfig, actualConfig)
	}
}
