package transfig_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/sironfoot/transfig"
)

type complex struct {
	StringValue string  `json:"stringValue"`
	IntValue    int     `json:"intValue"`
	FloatValue  float64 `json:"floatValue"`
	BoolValue   bool    `json:"boolValue"`

	SliceValueStrings []string  `json:"sliceValueStrings"`
	SliceValueFloats  []float64 `json:"sliceValueFloats"`
	SliceValueInts    []int     `json:"sliceValueInts"`
	SliceValueBools   []bool    `json:"sliceValueBools"`

	ObjectValue       subConfiguration `json:"objectValue"`
	SliceValueObjects []slicedConfig   `json:"sliceValueObjects"`
}

type subConfiguration struct {
	StringValue string              `json:"stringValue"`
	IntValue    int                 `json:"intValue"`
	ObjectValue subSubConfiguration `json:"objectValue"`
}

type subSubConfiguration struct {
	StringValue string `json:"stringValue"`
	IntValue    int    `json:"intValue"`
}

type slicedConfig struct {
	StringValue string `json:"stringValue"`
	IntValue    int    `json:"intValue"`
}

var expectedConfig = complex{
	StringValue: "Hello world",
	IntValue:    123,
	FloatValue:  123.45,
	BoolValue:   true,

	SliceValueStrings: []string{"string1", "string2", "string3"},
	SliceValueFloats:  []float64{1.2, 2.3, 3.4},
	SliceValueInts:    []int{1, 2, 3},
	SliceValueBools:   []bool{true, false, true},

	ObjectValue: subConfiguration{
		StringValue: "Hello world",
		IntValue:    123,

		ObjectValue: subSubConfiguration{
			StringValue: "Hello world",
			IntValue:    123,
		},
	},

	SliceValueObjects: []slicedConfig{
		slicedConfig{
			StringValue: "Hello world",
			IntValue:    123,
		},
		slicedConfig{
			StringValue: "Hello world",
			IntValue:    123,
		},
		slicedConfig{
			StringValue: "Hello world",
			IntValue:    123,
		},
	},
}

func TestLoad_WithEnvironment(t *testing.T) {
	// arrange
	var actualConfig complex

	altConfigString := `
    {
        "stringValue": "Hello world 2",
        "intValue": 456,
        "floatValue": 456.78,
        "boolValue": false,

        "sliceValueStrings": [ "string4", "string5", "string6", "string7", "string8" ],
		"sliceValueFloats": [ 4.5, 5.6, 7.8, 8.9 ],
		"sliceValueInts": [ 4, 5 ],
		"sliceValueBools": [ true ],

        "objectValue": {
            "stringValue": "Hello world 2",
            "intValue": 456,

            "objectValue": {
                "stringValue": "Hello world 2",
                "intValue": 456
            }
        },

        "sliceValueObjects": [
            {
                "stringValue": "Hello world 2",
                "intValue": 456
            },
            {
                "stringValue": "Hello world 2",
                "intValue": 456
            },
            {
                "stringValue": "Hello world 2",
                "intValue": 456
            },
            {
                "stringValue": "Hello world 2",
                "intValue": 456
            }
        ]
    }`

	var expectedAltConfig complex
	err := json.Unmarshal([]byte(altConfigString), &expectedAltConfig)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile("complex.test.json", []byte(altConfigString), 0644)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = os.Remove("complex.test.json")
		if err != nil {
			t.Fatal(err)
		}
	}()

	// act
	err = transfig.Load("complex.json", "test", &actualConfig)

	// assert
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expectedAltConfig, actualConfig) {
		t.Errorf("expected and actual config are different.\nExpected:\n%v\n\nActual:\n%v", expectedAltConfig, actualConfig)
	}
}

func TestLoad_WithPartialEnvironment(t *testing.T) {
	// arrange
	var actualConfig complex

	altConfigString := `
    {
        "stringValue": "Hello world 2",
		"sliceValueInts": [ 1, 2, 3, 4, 5, 6 ],

        "objectValue": {
            "stringValue": "Hello world 2",
            "objectValue": {
                "intValue": 456
            }
        }
    }`

	err := ioutil.WriteFile("complex.test.json", []byte(altConfigString), 0644)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = os.Remove("complex.test.json")
		if err != nil {
			t.Fatal(err)
		}
	}()

	// act
	err = transfig.Load("complex.json", "test", &actualConfig)

	// assert
	if err != nil {
		t.Fatal(err)
	}

	if actualConfig.StringValue != "Hello world 2" {
		t.Errorf("StringValue: expected '%s', actual '%s'", "Hello world 2", actualConfig.StringValue)
	}

	if len(actualConfig.SliceValueInts) != 6 {
		t.Errorf("SliceValueInts: expected %d items, actual was %d items", 6, len(actualConfig.SliceValueInts))
	}

	for i, num := range actualConfig.SliceValueInts {
		if num != (i + 1) {
			t.Errorf("SliceValueInts[%d]: expected %d, actual %d", i, i+1, num)
		}
	}

	if actualConfig.ObjectValue.StringValue != "Hello world 2" {
		t.Errorf("ObjectValue.StringValue: expected '%s', actual '%s'",
			"Hello world 2", actualConfig.ObjectValue.StringValue)
	}

	if actualConfig.ObjectValue.ObjectValue.IntValue != 456 {
		t.Errorf("ObjectValue.ObjectValue.StringValue: expected '%d', actual '%s'",
			456, actualConfig.ObjectValue.ObjectValue.StringValue)
	}

	// check other probs are same
	actualConfig.StringValue = expectedConfig.StringValue
	actualConfig.SliceValueInts = append(make([]int, 0), expectedConfig.SliceValueInts...)
	actualConfig.ObjectValue.StringValue = expectedConfig.ObjectValue.StringValue
	actualConfig.ObjectValue.ObjectValue.IntValue = expectedConfig.ObjectValue.ObjectValue.IntValue

	if !reflect.DeepEqual(expectedConfig, actualConfig) {
		t.Errorf("expected and actual config are different.\nExpected:\n%v\n\nActual:\n%v", expectedConfig, actualConfig)
	}
}
