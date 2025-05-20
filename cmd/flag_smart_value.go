package cmd

import (
	"fmt"
	"strconv"
)

type SmartValue struct {
	Value interface{}
}

func (s *SmartValue) String() string {
	return fmt.Sprintf("%v", s.Value)
}

func (s *SmartValue) Set(value string) error {
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		s.Value = intVal
		return nil
	}

	if boolVal, err := strconv.ParseBool(value); err == nil {
		s.Value = boolVal
		return nil
	}

	s.Value = value
	return nil
}

func (s *SmartValue) Type() string {
	return "smart"
}
