package utils

import (
	"strconv"
	"strings"
	"testing"
)

func AssertContains(t *testing.T, list []string, str string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}

	t.Errorf(t.Name() + " - Expected list: [" + strings.Join(list, ", ") + "] to contain: \"" + str + "\", but it did not.")
	return false
}

func AssertNotContains(t *testing.T, list []string, str string) bool {
	for _, s := range list {
		if s == str {
			t.Errorf(t.Name() + " - Expected list: [" + strings.Join(list, ", ") + "] to not contain: \"" + str + "\", but it did.")
			return false
		}
	}
	return true
}

func AssertStringEqual(t *testing.T, expected string, actual string) bool {
	if expected != actual {
		t.Errorf(t.Name() + " - Expected: \"" + expected + "\", but got: \"" + actual + "\"")
		return false
	}
	return true
}

// intentionally not using interface{}
func AssertEqual(t *testing.T, expected int, actual int) bool {
	if expected != actual {
		t.Errorf(t.Name() + " - Expected: \"" + strconv.Itoa(expected) + "\", but got: \"" + strconv.Itoa(actual) + "\"")
		return false
	}
	return true
}

func AssertMin(t *testing.T, min int, value int) bool {
	if value < min {
		t.Errorf(t.Name() + " - Expected: \"" + strconv.Itoa(value) + "\", to be at least: \"" + strconv.Itoa(min) + "\"")
		return false
	}
	return true
}

func AssertMax(t *testing.T, max int, value int) bool {
	if value > max {
		t.Errorf(t.Name() + " - Expected: \"" + strconv.Itoa(value) + "\", to be at most: \"" + strconv.Itoa(max) + "\"")
		return false
	}
	return true
}
