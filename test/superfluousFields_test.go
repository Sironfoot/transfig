package transfig_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/sironfoot/transfig"
)

type superfluousFields struct {
	StringValue string  `json:"stringValue"`
	IntValue    int     `json:"intValue"`
	FloatValue  float64 `json:"floatValue"`
	BoolValue   bool    `json:"boolValue"`
}

func TestLoad_SuperfluousFields(t *testing.T) {
	// arrange
	var actualConfig superfluousFields

	altConfigString := `
    {
        "stringValue": "Hello world 2",

        "nonsenseString": "Nonsense",
        "nonsenseInt": 123,
        "nonsenseBool": true,

        "nonsenceSlice": [ "nonsense1", "nonsense2" ],

        "nonsenseObject": {
            "nonsenseString": "Nonsense"
        }
    }`

	err := ioutil.WriteFile("superfluousFields.test.json", []byte(altConfigString), 0644)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = os.Remove("superfluousFields.test.json")
		if err != nil {
			t.Fatal(err)
		}
	}()

	// act
	err = transfig.Load("superfluousFields.json", "test", &actualConfig)

	// assert
	if err != nil {
		t.Fatal(err)
	}

	expected := superfluousFields{
		StringValue: "Hello world 2",
		FloatValue:  123.45,
		IntValue:    123,
		BoolValue:   true,
	}

	if !reflect.DeepEqual(expected, actualConfig) {
		t.Errorf("expected and actual config are different.\nExpected:\n%v\n\nActual:\n%v", expected, actualConfig)
	}
}
