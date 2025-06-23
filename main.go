package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	
	flag "github.com/spf13/pflag"
)

var versionString string

type Record struct {
	Key       string                 `json:"-"`
	KeyField  string                 `json:"-"`
	StartDate *time.Time             `json:"start_date"`
	EndDate   *time.Time             `json:"end_date"`
	Data      map[string]interface{} `json:"-"`
	DataOrder []string               `json:"-"`
}

func unmarshalRecords(data []byte, records *[]Record, keyField, startField, endField string) error {
	// Parse as raw JSON to get field order
	var rawArray []json.RawMessage
	if err := json.Unmarshal(data, &rawArray); err != nil {
		return err
	}

	*records = make([]Record, len(rawArray))
	for i, rawMsg := range rawArray {
		if err := unmarshalRecordFromJSON(rawMsg, &(*records)[i], keyField, startField, endField); err != nil {
			return err
		}
	}
	return nil
}

func unmarshalRecordFromJSON(data json.RawMessage, r *Record, keyField, startField, endField string) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Get ordered keys to find the first field (preserve original behavior)
	var orderedKeys []string
	if keyField == "" {
		decoder := json.NewDecoder(strings.NewReader(string(data)))
		decoder.Token() // {
		for decoder.More() {
			token, _ := decoder.Token()
			if key, ok := token.(string); ok {
				orderedKeys = append(orderedKeys, key)
				decoder.Token() // skip value
			}
		}
	}

	// Determine key field - use specified field or first field as default
	if keyField != "" {
		r.KeyField = keyField
	} else if len(orderedKeys) > 0 {
		r.KeyField = orderedKeys[0]
	} else {
		return fmt.Errorf("no fields found in record")
	}

	// Set key value
	if keyValue, ok := raw[r.KeyField]; ok {
		if keyStr, ok := keyValue.(string); ok {
			r.Key = keyStr
		} else {
			r.Key = fmt.Sprintf("%v", keyValue)
		}
	}

	// Handle start date field
	if startDateValue, ok := raw[startField]; ok {
		if startDateStr, ok := startDateValue.(string); ok && startDateStr != "" {
			startDate, err := parseDate(startDateStr)
			if err != nil {
				return fmt.Errorf("invalid %s format: %v", startField, err)
			}
			r.StartDate = &startDate
		}
	}

	// Handle end date field
	if endDateValue, ok := raw[endField]; ok {
		if endDateStr, ok := endDateValue.(string); ok && endDateStr != "" {
			endDate, err := parseDate(endDateStr)
			if err != nil {
				return fmt.Errorf("invalid %s format: %v", endField, err)
			}
			r.EndDate = &endDate
		}
	}

	// Store all other fields as data, preserving order
	r.Data = make(map[string]interface{})
	r.DataOrder = []string{}
	
	// Use original order if available, otherwise sort keys
	if len(orderedKeys) > 0 {
		for _, k := range orderedKeys {
			if k != r.KeyField {
				r.Data[k] = raw[k]
				r.DataOrder = append(r.DataOrder, k)
			}
		}
	} else {
		var allKeys []string
		for k := range raw {
			allKeys = append(allKeys, k)
		}
		for _, k := range allKeys {
			if k != r.KeyField {
				r.Data[k] = raw[k]
				r.DataOrder = append(r.DataOrder, k)
			}
		}
	}

	return nil
}


func main() {
	var (
		dateStr    = flag.StringP("date", "d", "", "Query date (YYYY-MM-DD or YYYYMMDD format, defaults to today)")
		keyField   = flag.StringP("key-field", "k", "", "Primary key field name (defaults to first field)")
		startField = flag.StringP("start-field", "s", "start_date", "Start date field name")
		endField   = flag.StringP("end-field", "e", "end_date", "End date field name")
		exitStatus = flag.BoolP("exit-status", "E", false, "Exit with non-zero status when no record found")
		version    = flag.BoolP("version", "v", false, "Show version information")
	)
	flag.CommandLine.SortFlags = false
	flag.Parse()

	if *version {
		if versionString == "" {
			fmt.Println("jdq (development version)")
		} else {
			fmt.Println(versionString)
		}
		return
	}

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: jdq [options] <key> <json-file>\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	queryKey := args[0]
	jsonFile := args[1]

	var queryDate time.Time
	var err error
	if *dateStr != "" {
		queryDate, err = parseDate(*dateStr)
		if err != nil {
			log.Fatalf("Invalid date format: %v", err)
		}
	} else {
		queryDate = time.Now()
	}

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	var records []Record
	if err := unmarshalRecords(data, &records, *keyField, *startField, *endField); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	// Check if date fields exist
	if len(records) > 0 {
		checkDateFields(records, *startField, *endField, jsonFile)
	}

	queryRecord(records, queryKey, queryDate, *exitStatus)
}

func checkDateFields(records []Record, startField, endField string, jsonFile string) {
	if len(records) == 0 {
		return
	}

	// We need to check the original JSON to see if the specified date fields exist
	// We'll re-parse just the first record to validate field existence
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return // If we can't read the file, skip validation
	}

	var rawRecords []map[string]interface{}
	if err := json.Unmarshal(data, &rawRecords); err != nil {
		return // If we can't parse, skip validation
	}

	if len(rawRecords) == 0 {
		return
	}

	// Check if the specified fields exist in the first record
	_, startFieldExists := rawRecords[0][startField]
	_, endFieldExists := rawRecords[0][endField]

	// Build error messages for missing fields
	var warnings []string
	if !startFieldExists {
		warnings = append(warnings, fmt.Sprintf("start date field '%s' not found in data", startField))
	}
	if !endFieldExists {
		warnings = append(warnings, fmt.Sprintf("end date field '%s' not found in data", endField))
	}

	if len(warnings) > 0 {
		log.Fatalf("Error: %s\nUse -s/-e to specify correct field names.", strings.Join(warnings, ", "))
	}
}

func queryRecord(records []Record, key string, date time.Time, exitOnMissing bool) {
	// CSS-style priority matching with value-level inheritance
	// Priority 1: Exact match (highest)
	// Priority 2: Default match (lower)
	// Within each priority level, later records override earlier ones
	
	var exactMatch *Record
	var defaultMatch *Record
	
	for i, record := range records {
		if isValidAt(record, date) {
			if record.Key == key {
				// Priority 1: Exact match - always override any previous exact match
				exactMatch = &records[i]
			} else if record.Key == "" {
				// Priority 2: Default match - only override previous default match
				defaultMatch = &records[i]
			}
		}
	}

	// Build final output with value-level inheritance, preserving field order
	output := make(map[string]interface{})
	var outputOrder []string
	
	// Determine which record to use for field ordering
	var orderSource *Record
	if exactMatch != nil {
		orderSource = exactMatch
	} else if defaultMatch != nil {
		orderSource = defaultMatch
	} else {
		// No match at all
		if exitOnMissing {
			fmt.Fprintf(os.Stderr, "No record found for key '%s' at date %s\n", key, date.Format("2006-01-02"))
			os.Exit(1)
		}
		// Return minimal object with just the key
		output := make(map[string]interface{})
		// Try to determine key field from any record
		keyField := "id"
		if len(records) > 0 {
			keyField = records[0].KeyField
		}
		output[keyField] = key
		jsonOutput, _ := json.Marshal(output)
		fmt.Println(string(jsonOutput))
		return
	}
	
	// Start with key field
	output[orderSource.KeyField] = getEffectiveKey(*orderSource, key)
	outputOrder = append(outputOrder, orderSource.KeyField)
	
	// Collect all possible fields from both records for complete inheritance
	allFields := make(map[string]bool)
	if defaultMatch != nil {
		for _, k := range defaultMatch.DataOrder {
			allFields[k] = true
		}
	}
	if exactMatch != nil {
		for _, k := range exactMatch.DataOrder {
			allFields[k] = true
		}
	}
	
	// Add fields in the order from the primary record, then any additional fields
	seenFields := make(map[string]bool)
	for _, k := range orderSource.DataOrder {
		if !seenFields[k] {
			// Apply inheritance for this field
			if defaultMatch != nil {
				if v, exists := defaultMatch.Data[k]; exists {
					output[k] = v
				}
			}
			if exactMatch != nil {
				if v, exists := exactMatch.Data[k]; exists {
					output[k] = v
				}
			}
			outputOrder = append(outputOrder, k)
			seenFields[k] = true
		}
	}
	
	// Add any remaining fields from the other record
	var otherRecord *Record
	if orderSource == exactMatch {
		otherRecord = defaultMatch
	} else {
		otherRecord = exactMatch
	}
	if otherRecord != nil {
		for _, k := range otherRecord.DataOrder {
			if !seenFields[k] {
				if v, exists := otherRecord.Data[k]; exists {
					output[k] = v
				}
				outputOrder = append(outputOrder, k)
				seenFields[k] = true
			}
		}
	}

	// Build ordered JSON output
	var result strings.Builder
	result.WriteString("{\n")
	for i, k := range outputOrder {
		if i > 0 {
			result.WriteString(",\n")
		}
		valueBytes, _ := json.Marshal(output[k])
		result.WriteString(fmt.Sprintf("  \"%s\": %s", k, string(valueBytes)))
	}
	result.WriteString("\n}")
	
	fmt.Println(result.String())
}

func parseDate(dateStr string) (time.Time, error) {
	// Try YYYYMMDD format first
	if len(dateStr) == 8 {
		if t, err := time.Parse("20060102", dateStr); err == nil {
			return t, nil
		}
	}
	// Fall back to YYYY-MM-DD format
	return time.Parse("2006-01-02", dateStr)
}

func isValidAt(record Record, date time.Time) bool {
	// If start_date is empty (nil), it matches any date (default behavior)
	if record.StartDate != nil && date.Before(*record.StartDate) {
		return false
	}
	// If end_date is empty (nil), it matches any date (default behavior)
	if record.EndDate != nil && date.After(*record.EndDate) {
		return false
	}
	return true
}


func getEffectiveKey(record Record, queryKey string) string {
	// If record key is empty (default), return the query key
	if record.Key == "" {
		return queryKey
	}
	return record.Key
}