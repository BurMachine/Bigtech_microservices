package configlib

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load загружает конфигурацию с приоритетом: ENV > YAML > defaults
func Load[T any](yamlPath string) (*T, error) {
	config := new(T)

	if err := setDefaults(config); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	if yamlPath != "" {
		if data, err := os.ReadFile(yamlPath); err == nil {
			// Файл существует и читается
			if err := yaml.Unmarshal(data, config); err != nil {
				return nil, fmt.Errorf("failed to parse yaml file %s: %w", yamlPath, err)
			}
		} else if !os.IsNotExist(err) {
			// Ошибка чтения файла (не "файл не найден")
			return nil, fmt.Errorf("failed to read yaml file %s: %w", yamlPath, err)
		}
	}

	if err := parseEnv(config, ""); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return config, nil
}

func setDefaults(config interface{}) error {
	return setDefaultsRecursive(reflect.ValueOf(config).Elem(), "")
}

func setDefaultsRecursive(val reflect.Value, prefix string) error {
	if val.Kind() != reflect.Struct {
		return nil
	}

	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := val.Field(i)

		if !field.IsExported() {
			continue
		}

		envPrefix := field.Tag.Get("envPrefix")
		newPrefix := prefix
		if envPrefix != "" {
			newPrefix = prefix + envPrefix
		}

		defaultVal := field.Tag.Get("envDefault")
		if defaultVal != "" && fieldVal.CanSet() {
			if err := setFieldValue(fieldVal, defaultVal); err != nil {
				return fmt.Errorf("failed to set default for field %s: %w", field.Name, err)
			}
		}

		if field.Type.Kind() == reflect.Struct {
			if err := setDefaultsRecursive(fieldVal, newPrefix); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseEnv(config interface{}, prefix string) error {
	return parseEnvRecursive(reflect.ValueOf(config).Elem(), prefix)
}

func parseEnvRecursive(val reflect.Value, prefix string) error {
	if val.Kind() != reflect.Struct {
		return nil
	}

	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := val.Field(i)

		if !field.IsExported() {
			continue
		}

		// Получаем имя ENV переменной
		envName := field.Tag.Get("env")
		if envName == "" && field.Type.Kind() != reflect.Struct {
			// Если нет тега env и это не структура - пропускаем
			continue
		}

		// Получаем envPrefix для вложенных структур
		envPrefix := field.Tag.Get("envPrefix")
		newPrefix := prefix
		if envPrefix != "" {
			newPrefix = prefix + envPrefix
		}

		// Пытаемся прочитать значение из ENV
		if envName != "" {
			fullEnvName := prefix + envName
			if envValue, exists := os.LookupEnv(fullEnvName); exists && fieldVal.CanSet() {
				if err := setFieldValue(fieldVal, envValue); err != nil {
					return fmt.Errorf("failed to set env value for %s: %w", fullEnvName, err)
				}
			}
		}

		// Рекурсия для вложенных структур
		if field.Type.Kind() == reflect.Struct {
			if err := parseEnvRecursive(fieldVal, newPrefix); err != nil {
				return err
			}
		}
	}
	return nil
}

// setFieldValue устанавливает значение поля из строки
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Slice:
		// Поддержка слайсов (разделитель - запятая)
		if field.Type().Elem().Kind() == reflect.String {
			field.Set(reflect.ValueOf(strings.Split(value, ",")))
		}
	default:
		return fmt.Errorf("unsupported type: %s", field.Kind())
	}
	return nil
}

// LoadWithValidation загружает конфигурацию и валидирует её
func LoadWithValidation[T any](yamlPath string, validate func(*T) error) (*T, error) {
	config, err := Load[T](yamlPath)
	if err != nil {
		return nil, err
	}

	if validate != nil {
		if err := validate(config); err != nil {
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	}

	return config, nil
}
