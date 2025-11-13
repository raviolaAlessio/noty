package flags

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// choiceValue implements the [pflag.Value] interface.
type choiceValue[T any] struct {
	value     T
	validate  func(T) error
	convert   func(string) (T, error)
	toString  func(T) string
	valueType string
}

// Set sets the value of the choice.
func (f *choiceValue[T]) Set(s string) error {
	v, err := f.convert(s)
	if err != nil {
		return err
	}
	err = f.validate(v)
	if err != nil {
		return err
	}

	f.value = v
	return nil
}

// Type returns the type of the choice
func (f *choiceValue[T]) Type() string { return f.valueType }

// String returns the current value of the choice.
func (f *choiceValue[T]) String() string { return f.toString(f.value) }

func StringChoice(choices []string, defaultValue string) *choiceValue[string] {
	return &choiceValue[string]{
		value: defaultValue,
		validate: func(s string) error {
			if slices.Contains(choices, s) {
				return nil
			}
			return fmt.Errorf("must be one of %v", choices)
		},
		convert:   func(s string) (string, error) { return s, nil },
		toString:  func(s string) string { return s },
		valueType: "string",
	}
}

func StringChoiceOrInt(choices []string, defaultValue string) *choiceValue[string] {
	return &choiceValue[string]{
		value: defaultValue,
		validate: func(s string) error {
			if slices.Contains(choices, s) {
				return nil
			}
			if _, err := strconv.Atoi(s); err == nil {
				return nil
			}
			return fmt.Errorf("must be one of %v or an integer", choices)
		},
		convert:   func(s string) (string, error) { return s, nil },
		toString:  func(s string) string { return s },
		valueType: "string",
	}
}

func NumberChoice(choices []int, defaultValue int) *choiceValue[int] {
	return &choiceValue[int]{
		value: defaultValue,
		validate: func(s int) error {
			if slices.Contains(choices, s) {
				return nil
			}
			return fmt.Errorf("must be one of %v", choices)
		},
		convert:   func(s string) (int, error) { return strconv.Atoi(s) },
		toString:  func(i int) string { return fmt.Sprintf("%d", i) },
		valueType: "int",
	}
}

// choiceValue implements the [pflag.Value] interface.
type choiceValueSlice[T any] struct {
	value     *[]T
	validate  func(*[]T) error
	convert   func(string) (*[]T, error)
	toString  func(*[]T) string
	valueType string
}

// Set sets the value of the choice.
func (f *choiceValueSlice[T]) Set(s string) error {
	v, err := f.convert(s)
	if err != nil {
		return err
	}
	err = f.validate(v)
	if err != nil {
		return err
	}

	f.value = v
	return nil
}

func (f *choiceValueSlice[T]) Type() string { return f.valueType }

func (f *choiceValueSlice[T]) String() string { return f.toString(f.value) }

func (s *choiceValueSlice[T]) Append(val T) error {
	*s.value = append(*s.value, val)
	return nil
}

func (f *choiceValueSlice[T]) Replace(val []T) error {
	*f.value = val
	return nil
}

func (s *choiceValueSlice[T]) GetSlice() []T { return *s.value }

func writeAsCSV(vals []string) (string, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	err := w.Write(vals)
	if err != nil {
		return "", err
	}
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n"), nil
}

func StringChoiceSlice(choices []string, defaultValue []string) *choiceValueSlice[string] {
	return &choiceValueSlice[string]{
		value: &defaultValue,
		validate: func(s *[]string) error {
			if s == nil || len(*s) == 0 {
				return nil
			}
			for _, v := range *s {
				if !slices.Contains(choices, v) {
					return fmt.Errorf("%s not valid, must be one of %v", v, choices)
				}
			}
			return nil
		},
		convert: func(s string) (*[]string, error) {
			if len(s) == 0 {
				return &[]string{}, nil
			}
			val := strings.Split(s, ",")
			return &val, nil
		},
		toString: func(s *[]string) string {
			str, _ := writeAsCSV(*s)
			return "[" + str + "]"
		},
		valueType: "stringSlice",
	}
}
