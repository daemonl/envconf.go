package envconf

import (
	"fmt"
	"os"
	"reflect"
)

func Parse(dest interface{}) error {
	rt := reflect.TypeOf(dest).Elem()
	rv := reflect.ValueOf(dest).Elem()
	for i := 0; i < rv.NumField(); i++ {
		tag := rt.Field(i).Tag
		envName := tag.Get("env")
		envVal := os.Getenv(envName)
		if envVal == "" {
			defaultValue := tag.Get("default")
			if defaultValue == "" {

				return fmt.Errorf("Required ENV var not set: %v", tag)
			}
			envVal = defaultValue
		}
		rv.Field(i).SetString(envVal)
	}
	return nil
}
