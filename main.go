// Package transfig provides utilities for loading a JSON configuration file into a struct object
// graph, with support for providing an alternative environment JSON config file
// (e.g. "dev", "staging", "uat", "live"), with values replaced using transformations. Inspired by
// the way Microsoft ASP.NET handles configuration files.
package transfig

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

// ErrPrimaryConfigFileNotExist is returned if the primary JSON config file doesn't exist
var ErrPrimaryConfigFileNotExist = fmt.Errorf("config: primary config file does not exist")

// ErrConfigDataNotPointer is returned when the configData
// struct to pass config data into is not a pointer
var ErrConfigDataNotPointer = fmt.Errorf("config: configData argument is not a pointer")

type configFile struct {
	Path                string
	PathLastModified    time.Time
	AltExists           bool
	AltPath             string
	AltPathLastModified time.Time
	ConfigData          interface{}
}

var (
	cache    = make(map[string]configFile)
	cacheMux = sync.RWMutex{}
)

var (
	// ReloadPollingInterval sets how often config files are checked for changes
	ReloadPollingInterval = time.Duration(time.Second * 5)
	pollingIntervalMutex  = sync.Mutex{}
	stopPolling           = make(chan bool)
	pollingTicker         *time.Ticker
)

// SetReloadPollingInterval determines how often we should
// check for changes to the config files, so we can reload them.
func SetReloadPollingInterval(duration time.Duration) {
	pollingIntervalMutex.Lock()
	defer pollingIntervalMutex.Unlock()

	ReloadPollingInterval = duration

	if pollingTicker != nil {
		stopPolling <- true
		pollingTicker.Stop()
		pollingTicker = nil
	}

	pollingTicker = time.NewTicker(ReloadPollingInterval)

	go func(ticker *time.Ticker) {
		for {
			select {
			case <-ticker.C:
				for key, config := range cache {
					pathInfo, err := os.Stat(config.Path)
					if err != nil {
						break
					}

					var altPathInfo os.FileInfo
					if config.AltExists {
						altPathInfo, err = os.Stat(config.AltPath)
						if err != nil {
							break
						}
					}

					if pathInfo.ModTime().After(config.PathLastModified) ||
						(config.AltExists && altPathInfo.ModTime().After(config.AltPathLastModified)) {
						cacheMux.Lock()
						delete(cache, key)
						cacheMux.Unlock()
					}
				}
			case <-stopPolling:
				return
			}
		}
	}(pollingTicker)
}

func init() {
	SetReloadPollingInterval(time.Duration(time.Second * 5))
}

// LoadWithCaching will load a configuration json file into a struct with built in support for caching
func LoadWithCaching(path, environment string, configData interface{}) error {
	cacheMux.RLock()

	cacheKey := path + "_" + environment
	cachedConfig, isCached := cache[cacheKey]

	if isCached {
		value := reflect.ValueOf(cachedConfig.ConfigData)
		reflect.ValueOf(configData).Elem().Set(value)
		cacheMux.RUnlock()
		return nil
	}

	cacheMux.RUnlock()

	cacheMux.Lock()
	defer cacheMux.Unlock()

	cachedConfig, isCached = cache[cacheKey]
	if isCached {
		value := reflect.ValueOf(cachedConfig.ConfigData)
		reflect.ValueOf(configData).Elem().Set(value)
		return nil
	}

	err := Load(path, environment, configData)
	if err != nil {
		return err
	}

	pathInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	altExists := true
	altPath := generateAltPath(path, environment)
	altPathInfo, err := os.Stat(altPath)
	if os.IsNotExist(err) {
		altExists = false
	} else if err != nil {
		return err
	}

	configInfo := configFile{
		Path:             path,
		PathLastModified: pathInfo.ModTime(),
		AltExists:        altExists,
		ConfigData:       reflect.ValueOf(configData).Elem().Interface(),
	}

	if altExists {
		configInfo.AltPath = altPath
		configInfo.AltPathLastModified = altPathInfo.ModTime()
	}

	cache[cacheKey] = configInfo

	return nil
}

// Load will load a configuration json file into a struct
func Load(path, environment string, configData interface{}) (err error) {
	if reflect.TypeOf(configData).Kind() != reflect.Ptr {
		return ErrConfigDataNotPointer
	}

	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return ErrPrimaryConfigFileNotExist
	} else if err != nil {
		return fmt.Errorf("config: error opening primary config file: %s", err)
	}

	err = json.NewDecoder(file).Decode(configData)
	if err != nil {
		return fmt.Errorf("config: cannot unmarshal config file: %s", err)
	}

	altPath := generateAltPath(path, environment)

	altFile, err := os.Open(altPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("config: error opening environment config file \"%s\": %s", altPath, err)
	}

	if os.IsNotExist(err) {
		return nil
	}

	altConfigData := map[string]interface{}{}
	err = json.NewDecoder(altFile).Decode(&altConfigData)
	if err != nil {
		return fmt.Errorf("config: cannot unmarshal environment config file: %s", err)
	}

	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("config: parsing environment config file: %s", p)
		}
	}()

	configValue := reflect.ValueOf(configData).Elem()
	parseMap(altConfigData, configValue)

	return
}

func generateAltPath(path, environment string) string {
	return strings.Replace(path, ".json", "."+environment+".json", 1)
}

func parseMap(aMap map[string]interface{}, configValue reflect.Value) {
	for key, value := range aMap {
		fieldName := ""

		for i := 0; i < configValue.NumField(); i++ {
			fieldInfo := configValue.Type().Field(i)
			jsonFieldName := strings.TrimSpace(fieldInfo.Tag.Get("json"))

			if jsonFieldName == key {
				fieldName = fieldInfo.Name
			}
		}

		fieldValue := configValue.FieldByName(fieldName)
		if fieldValue.Kind() == reflect.Invalid {
			continue
		}

		switch realValue := value.(type) {
		case map[string]interface{}:
			if fieldValue.Kind() == reflect.Struct || fieldValue.Kind() == reflect.Map {
				parseMap(realValue, fieldValue)
			}
		case []interface{}:
			if fieldValue.Kind() == reflect.Slice {
				parseSlice(realValue, fieldValue)
			}
		case string:
			if fieldValue.Kind() == reflect.String {
				fieldValue.SetString(realValue)
			}
		case float64:
			switch fieldValue.Kind() {
			case reflect.Float32, reflect.Float64:
				fieldValue.SetFloat(realValue)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fieldValue.SetInt(int64(realValue))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				fieldValue.SetUint(uint64(realValue))
			}
		case bool:
			if fieldValue.Kind() == reflect.Bool {
				fieldValue.SetBool(realValue)
			}
		}
	}
}

func parseSlice(aSlice []interface{}, configValue reflect.Value) {
	newSlice := reflect.MakeSlice(configValue.Type(), len(aSlice), len(aSlice))
	configValue.Set(newSlice)

	for i, value := range aSlice {
		configItem := configValue.Index(i)

		switch realItem := value.(type) {
		case map[string]interface{}:
			if configItem.Kind() == reflect.Struct || configItem.Kind() == reflect.Map {
				parseMap(realItem, configItem)
			}
		case []interface{}:
			if configItem.Kind() == reflect.Slice {
				parseSlice(realItem, configItem)
			}
		case string:
			if configItem.Kind() == reflect.String {
				configItem.SetString(realItem)
			}
		case float64:
			switch configItem.Kind() {
			case reflect.Float32, reflect.Float64:
				configItem.SetFloat(realItem)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				configItem.SetInt(int64(realItem))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				configItem.SetUint(uint64(realItem))
			}
		case bool:
			if configItem.Kind() == reflect.Bool {
				configItem.SetBool(realItem)
			}
		}
	}
}
