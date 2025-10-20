package utils

import (
	"regexp"
	"testing"
)

func Test_GenerateId_Format(t *testing.T) {
	// arrange
	id := GenerateId()
	expectedPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}-\d+$`

	// act
	matched, err := regexp.MatchString(expectedPattern, id)

	// assert
	if err != nil {
		t.Fatalf("Error compiling regex: %v", err)
	}

	if !matched {
		t.Errorf("GenerateId() returned invalid format: %s", id)
	}
}

func Test_GenerateId_Uniqueness(t *testing.T) {
	// act
	id1 := GenerateId()
	id2 := GenerateId()

	// assert
	if id1 == id2 {
		t.Errorf("GenerateId() produced duplicate IDs: %s", id1)
	}
}

func Test_GetCurrentTimeStamp_FormatWithoutSeconds(t *testing.T) {
	// arrange
	expectedPattern := `^\d{4}-\d{2}-\d{2} \d{2}_\d{2}$`

	// act
	timestamp := GetCurrentTimeStamp(false)

	// assert
	matched, err := regexp.MatchString(expectedPattern, timestamp)
	if err != nil {
		t.Fatalf("Error compiling regex: %v", err)
	}

	if !matched {
		t.Errorf("GetCurrentTimeStamp(false) returned invalid format: %s", timestamp)
	}
}

func Test_GetCurrentTimeStamp_FormatWithSeconds(t *testing.T) {
	// arrange
	expectedPattern := `^\d{4}-\d{2}-\d{2} \d{2}_\d{2}_\d{2}$`

	// act
	timestamp := GetCurrentTimeStamp(true)

	// assert
	matched, err := regexp.MatchString(expectedPattern, timestamp)
	if err != nil {
		t.Fatalf("Error compiling regex: %v", err)
	}

	if !matched {
		t.Errorf("GetCurrentTimeStamp(true) returned invalid format: %s", timestamp)
	}
}
