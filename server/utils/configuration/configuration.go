// Package configuration provides a single generic way to populate a config
// struct from environment variables, driven entirely by struct tags:
//
//	type HttpConfig struct {
//		Host string `env:"HTTP_HOST" bind:"required"`
//		Port int    `env:"HTTP_PORT" bind:"optional" default:"8000"`
//	}
//
// `env` names the environment variable. `bind` is "required" (Load fails if
// unset) or "optional" (falls back to `default`, or the field's zero value
// if there's no `default` tag). Fields with no `env` tag are left alone —
// useful for struct-local helper methods that don't come from the
// environment.
package configuration

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Load builds a new T, populating every `env`-tagged field from the
// process environment. T must be a struct type.
func Load[T any]() (T, error) {
	var cfg T
	v := reflect.ValueOf(&cfg).Elem()
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return cfg, fmt.Errorf("configuration: %s is not a struct", t.Kind())
	}

	var missing []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		envKey, bound := field.Tag.Lookup("env")
		if !bound {
			continue
		}

		raw, present := os.LookupEnv(envKey)
		if !present {
			if field.Tag.Get("bind") == "required" {
				missing = append(missing, envKey)
				continue
			}
			def, hasDefault := field.Tag.Lookup("default")
			if !hasDefault {
				continue
			}
			raw = def
		}

		if err := setField(v.Field(i), raw); err != nil {
			return cfg, fmt.Errorf("configuration: invalid value for %s: %w", envKey, err)
		}
	}

	if len(missing) > 0 {
		return cfg, fmt.Errorf("configuration: missing required environment variable(s): %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

func setField(field reflect.Value, raw string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("parsing %q as int: %w", raw, err)
		}
		field.SetInt(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("parsing %q as bool: %w", raw, err)
		}
		field.SetBool(b)
	default:
		return fmt.Errorf("unsupported field type %s", field.Kind())
	}
	return nil
}
