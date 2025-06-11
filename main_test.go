package main

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"20240522", "2024-05-22", false},
		{"2024-05-22", "2024-05-22", false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, test := range tests {
		result, err := parseDate(test.input)
		
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %q, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", test.input, err)
				continue
			}
			
			expected, _ := time.Parse("2006-01-02", test.expected)
			if !result.Equal(expected) {
				t.Errorf("For input %q, expected %v, got %v", test.input, expected, result)
			}
		}
	}
}

func TestIsValidAt(t *testing.T) {
	// Test date
	testDate, _ := time.Parse("2006-01-02", "2024-05-22")
	
	tests := []struct {
		name     string
		record   Record
		expected bool
	}{
		{
			name: "No date restrictions (both nil)",
			record: Record{
				StartDate: nil,
				EndDate:   nil,
			},
			expected: true,
		},
		{
			name: "Within date range",
			record: Record{
				StartDate: parseTimePtr("2024-05-01"),
				EndDate:   parseTimePtr("2024-05-31"),
			},
			expected: true,
		},
		{
			name: "Before range",
			record: Record{
				StartDate: parseTimePtr("2024-06-01"),
				EndDate:   parseTimePtr("2024-06-30"),
			},
			expected: false,
		},
		{
			name: "After range",
			record: Record{
				StartDate: parseTimePtr("2024-04-01"),
				EndDate:   parseTimePtr("2024-04-30"),
			},
			expected: false,
		},
		{
			name: "Only start_date set (within)",
			record: Record{
				StartDate: parseTimePtr("2024-05-01"),
				EndDate:   nil,
			},
			expected: true,
		},
		{
			name: "Only end_date set (within)",
			record: Record{
				StartDate: nil,
				EndDate:   parseTimePtr("2024-05-31"),
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isValidAt(test.record, testDate)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}


func TestGetEffectiveKey(t *testing.T) {
	tests := []struct {
		name     string
		record   Record
		queryKey string
		expected string
	}{
		{
			name:     "Return record key when not empty",
			record:   Record{Key: "1001"},
			queryKey: "1001",
			expected: "1001",
		},
		{
			name:     "Return query key when record key is empty",
			record:   Record{Key: ""},
			queryKey: "9999",
			expected: "9999",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getEffectiveKey(test.record, test.queryKey)
			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

// Helper function to create time pointers
func parseTimePtr(dateStr string) *time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return &t
}