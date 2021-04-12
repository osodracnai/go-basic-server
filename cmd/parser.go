package cmd

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"net/url"
	"reflect"
	"strings"
	"time"
)

func DecoderConfigOptions(config *mapstructure.DecoderConfig) {
	config.DecodeHook = mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		mapstructure.StringToTimeHookFunc(time.Kitchen),
		StringToURL(),
	)
}

func StringToURL() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Struct {
			return data, nil
		}

		dataVal := data.(string)

		// Add network interface if not provided
		if strings.HasPrefix(dataVal, ":") {
			dataVal = fmt.Sprintf("127.0.0.1%s", dataVal)
		}

		// Add schema if not provided
		if !(strings.HasPrefix(dataVal, "http://") || strings.HasPrefix(dataVal, "https://")) {
			dataVal = fmt.Sprintf("http://%s", dataVal)
		}

		listenURL, err := url.Parse(dataVal)
		if err != nil {
			return nil, fmt.Errorf("parse address: %v", err)
		}
		return listenURL, nil
	}
}
