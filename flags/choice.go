package flags

import (
	"fmt"
	"slices"
	"strconv"
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

func StringChoice(choices []string) *choiceValue[string] {
	return &choiceValue[string]{
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

func StringChoiceOrNumber(choices []string) *choiceValue[string] {
	return &choiceValue[string]{
		validate: func(s string) error {
			if slices.Contains(choices, s) {
				return nil
			}
			if _, err := strconv.Atoi(s); err == nil {
				return nil
			}
			return fmt.Errorf("must be one of %v", choices)
		},
		convert:   func(s string) (string, error) { return s, nil },
		toString:  func(s string) string { return s },
		valueType: "string",
	}
}

func NumberChoice(choices []int) *choiceValue[int] {
	return &choiceValue[int]{
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
