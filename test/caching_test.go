package transfig_test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

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

func TestCaching(t *testing.T) {
	defaultDuration := transfig.ReloadPollingInterval
	defer func() {
		transfig.SetReloadPollingInterval(defaultDuration)
	}()

	// arrange
	transfig.SetReloadPollingInterval(time.Duration(time.Second * 1))

	err := copyFile("_complexCachingTest.json", "complex.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.Remove("_complexCachingTest.json")
		if err != nil {
			t.Fatal(err)
		}
	}()
	<-time.After(time.Duration(time.Second * 1))

	// act
	var configDataCopy complex
	err = transfig.LoadWithCaching("_complexCachingTest.json", "test", &configDataCopy)
	if err != nil {
		t.Fatal(err)
	}

	originalValue := configDataCopy.IntValue
	newValue := originalValue + 1

	configDataCopy.IntValue = newValue

	data, err := json.Marshal(&configDataCopy)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile("_complexCachingTest.json", data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// assert
	var preCacheConfigData complex
	err = transfig.LoadWithCaching("_complexCachingTest.json", "test", &preCacheConfigData)
	if err != nil {
		t.Fatal(err)
	}

	if preCacheConfigData.IntValue != originalValue {
		t.Errorf("pre-caching: expected: %d but got %d", originalValue, preCacheConfigData.IntValue)
	}

	<-time.After(time.Duration(time.Second * 1))

	var postCacheConfigData complex
	err = transfig.LoadWithCaching("_complexCachingTest.json", "test", &postCacheConfigData)
	if err != nil {
		t.Fatal(err)
	}

	if postCacheConfigData.IntValue != newValue {
		t.Errorf("post-caching: expected: %d but got %d", newValue, postCacheConfigData.IntValue)
	}
}

func TestCachedVersionsAreCopies(t *testing.T) {
	// arrange
	var complex1 complex
	err := transfig.LoadWithCaching("complex.json", "test", &complex1)
	if err != nil {
		t.Fatal(err)
	}

	var complex2 complex
	err = transfig.LoadWithCaching("complex.json", "test", &complex2)
	if err != nil {
		t.Fatal(err)
	}

	// act/assert
	if complex1.IntValue != complex2.IntValue {
		t.Errorf("complex1 value (%d) should be the same as complex2 value (%d)", complex1.IntValue, complex2.IntValue)
	}

	complex1.IntValue += 1

	if complex1.IntValue == complex2.IntValue {
		t.Errorf("complex1 value (%d) should be different to complex2 value (%d)", complex1.IntValue, complex2.IntValue)
	}
}
