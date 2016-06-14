package transfig_test

import (
	"os"
	"testing"
	"time"

	"github.com/sironfoot/transfig"
)

func BenchmarkLoad(b *testing.B) {
	err := copyFile("complex.test.json", "complex.json")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		err = os.Remove("complex.test.json")
		if err != nil {
			b.Fatal(err)
		}
	}()

	for n := 0; n < b.N; n++ {
		var complexData complex
		err := transfig.Load("complex.json", "test", &complexData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoadWithCaching(b *testing.B) {
	defaultDuration := transfig.ReloadPollingInterval
	defer func() {
		transfig.SetReloadPollingInterval(defaultDuration)
	}()

	// arrange
	transfig.SetReloadPollingInterval(time.Duration(time.Second * 1))

	err := copyFile("complex.test.json", "complex.json")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		err = os.Remove("complex.test.json")
		if err != nil {
			b.Fatal(err)
		}
	}()

	for n := 0; n < b.N; n++ {
		var complexData complex
		err := transfig.LoadWithCaching("complex.json", "test", &complexData)
		if err != nil {
			b.Fatal(err)
		}
	}
}
