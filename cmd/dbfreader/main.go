package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	dbf "github.com/SebastiaanKlippert/go-foxpro-dbf"
)

func main() {
	// Check if DBF file is provided as argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <DBF_FILE> [ENCODING] [OPTIONS]")
		fmt.Println("Example: go run main.go ../../testdata/TEST.DBF")
		fmt.Println("Example: go run main.go myfile.dbf big5")
		fmt.Println("Example: go run main.go myfile.dbf big5 --csv")
		fmt.Println("Supported encodings: win1250 (default), big5, utf8")
		fmt.Println("Options:")
		fmt.Println("  --csv          Export to CSV file (same name as DBF)")
		fmt.Println("  --csv=file.csv Export to specified CSV file")
		fmt.Println("  --no-display   Skip console display (useful with --csv)")
		os.Exit(1)
	}

	dbfFile := os.Args[1]
	encoding := "big5" // default
	csvOutput := ""
	noDisplay := false

	// Parse arguments
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--csv") {
			if arg == "--csv" {
				// Generate CSV filename from DBF filename
				csvOutput = strings.TrimSuffix(dbfFile, filepath.Ext(dbfFile)) + ".csv"
			} else if strings.HasPrefix(arg, "--csv=") {
				csvOutput = strings.TrimPrefix(arg, "--csv=")
			}
		} else if arg == "--no-display" {
			noDisplay = true
		} else if i == 2 && !strings.HasPrefix(arg, "--") {
			// Second argument is encoding if it doesn't start with --
			encoding = strings.ToLower(arg)
		}
	}

	if !noDisplay {
		fmt.Printf("Opening DBF file: %s (encoding: %s)\n", dbfFile, encoding)
		if csvOutput != "" {
			fmt.Printf("Will export to CSV: %s\n", csvOutput)
		}
	}

	// Choose decoder based on encoding parameter
	var decoder dbf.Decoder
	switch encoding {
	case "big5":
		decoder = new(dbf.Big5Decoder)
	case "utf8":
		decoder = new(dbf.UTF8Decoder)
	case "win1250":
		decoder = new(dbf.Win1250Decoder)
	default:
		fmt.Printf("Unsupported encoding: %s. Using win1250 as default.\n", encoding)
		decoder = new(dbf.Win1250Decoder)
	}

	// Open the DBF file with the chosen decoder
	d, err := dbf.OpenFile(dbfFile, decoder)
	if err != nil {
		log.Fatalf("Error opening DBF file: %v", err)
	}
	defer d.Close()

	// Print basic file information
	if !noDisplay {
		fmt.Printf("Total records: %d\n", d.NumRecords())
		fmt.Printf("Number of fields: %d\n", d.NumFields())
		fmt.Println("Field names:", d.FieldNames())
	}

	// Export to CSV if requested
	if csvOutput != "" {
		err = exportToCSV(d, csvOutput, noDisplay)
		if err != nil {
			log.Fatalf("Error exporting to CSV: %v", err)
		}
		if !noDisplay {
			fmt.Printf("Successfully exported %d records to %s\n", d.NumRecords(), csvOutput)
		}
		if noDisplay {
			return // Exit early if no display requested
		}
	}

	// Print field information
	if !noDisplay {
		fmt.Println("\nField details:")
		for i, field := range d.Fields() {
			fmt.Printf("  %d: %s (Type: %s, Length: %d, Decimals: %d)\n",
				i, field.FieldName(), field.FieldType(), field.Len, field.Decimals)
		}

		// Read and display first few records (up to 10)
		maxRecords := uint32(10)
		if d.NumRecords() < maxRecords {
			maxRecords = d.NumRecords()
		}

		fmt.Printf("\nFirst %d records:\n", maxRecords)
		for i := uint32(0); i < maxRecords; i++ {
			record, err := d.RecordAt(i)
			if err != nil {
				log.Printf("Error reading record %d: %v", i, err)
				continue
			}

			// Skip deleted records
			if record.Deleted {
				fmt.Printf("Record %d: [DELETED]\n", i)
				continue
			}

			// Print record with field names and values
			fmt.Printf("Record %d:\n", i)
			fieldNames := d.FieldNames()
			fieldSlice := record.FieldSlice()

			for j, value := range fieldSlice {
				if j < len(fieldNames) {
					// Use cast helpers for cleaner display
					switch d.Fields()[j].FieldType() {
					case "C": // Character
						fmt.Printf("  %s: %q\n", fieldNames[j], dbf.ToTrimmedString(value))
					case "N", "F": // Numeric, Float
						if d.Fields()[j].Decimals == 0 {
							fmt.Printf("  %s: %d\n", fieldNames[j], dbf.ToInt64(value))
						} else {
							fmt.Printf("  %s: %.2f\n", fieldNames[j], dbf.ToFloat64(value))
						}
					case "L": // Logical
						fmt.Printf("  %s: %t\n", fieldNames[j], dbf.ToBool(value))
					case "D", "T": // Date, DateTime
						fmt.Printf("  %s: %v\n", fieldNames[j], dbf.ToTime(value))
					default:
						fmt.Printf("  %s: %v\n", fieldNames[j], value)
					}
				}
			}
			fmt.Println()
		}
	}

	// Demonstrate record navigation
	if !noDisplay && d.NumRecords() > 0 {
		fmt.Println("Demonstrating record navigation:")

		// Go to first record
		err = d.GoTo(0)
		if err != nil {
			log.Printf("Error going to first record: %v", err)
		} else {
			fmt.Printf("At record 0, EOF: %t, BOF: %t\n", d.EOF(), d.BOF())

			// Read specific field by position
			if d.NumFields() > 0 {
				field0, err := d.Field(0)
				if err != nil {
					log.Printf("Error reading field 0: %v", err)
				} else {
					fmt.Printf("Field 0 value: %v\n", dbf.ToTrimmedString(field0))
				}
			}

			// Skip to next record
			d.Skip(1)
			fmt.Printf("After Skip(1), at record pointer, EOF: %t\n", d.EOF())
		}
	}

	// Convert a record to JSON (if records exist)
	if !noDisplay && d.NumRecords() > 0 {
		fmt.Println("\nRecord 0 as JSON:")
		jsonData, err := d.RecordToJSON(0, true) // trim spaces
		if err != nil {
			log.Printf("Error converting to JSON: %v", err)
		} else {
			fmt.Printf("%s\n", jsonData)
		}
	}

	if !noDisplay {
		fmt.Println("\nDone!")
	}
}

// exportToCSV exports all records from the DBF to a CSV file
func exportToCSV(d *dbf.DBF, filename string, silent bool) error {
	if !silent {
		fmt.Printf("Exporting to CSV: %s...\n", filename)
	}

	// Create CSV file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header row (field names)
	fieldNames := d.FieldNames()
	if err := writer.Write(fieldNames); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data rows
	totalRecords := d.NumRecords()
	processedRecords := uint32(0)

	for i := uint32(0); i < totalRecords; i++ {
		record, err := d.RecordAt(i)
		if err != nil {
			if !silent {
				log.Printf("Error reading record %d: %v", i, err)
			}
			continue
		}

		// Skip deleted records
		if record.Deleted {
			continue
		}

		// Convert record to string slice
		fieldSlice := record.FieldSlice()
		csvRow := make([]string, len(fieldSlice))

		for j, value := range fieldSlice {
			csvRow[j] = formatValueForCSV(value, d.Fields()[j])
		}

		if err := writer.Write(csvRow); err != nil {
			return fmt.Errorf("failed to write CSV row %d: %v", i, err)
		}

		processedRecords++

		// Show progress for large files
		if !silent && totalRecords > 1000 && processedRecords%1000 == 0 {
			fmt.Printf("Processed %d/%d records...\n", processedRecords, totalRecords)
		}
	}

	return nil
}

// formatValueForCSV formats a field value for CSV output
func formatValueForCSV(value interface{}, field dbf.FieldHeader) string {
	if value == nil {
		return ""
	}

	switch field.FieldType() {
	case "C": // Character
		return dbf.ToTrimmedString(value)
	case "N": // Numeric
		if field.Decimals == 0 {
			return strconv.FormatInt(dbf.ToInt64(value), 10)
		}
		return strconv.FormatFloat(dbf.ToFloat64(value), 'f', int(field.Decimals), 64)
	case "F": // Float
		return strconv.FormatFloat(dbf.ToFloat64(value), 'f', -1, 64)
	case "L": // Logical
		if dbf.ToBool(value) {
			return "true"
		}
		return "false"
	case "D": // Date
		t := dbf.ToTime(value)
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02")
	case "T": // DateTime
		t := dbf.ToTime(value)
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02 15:04:05")
	case "M": // Memo
		return dbf.ToString(value)
	case "I": // Integer
		return strconv.FormatInt(dbf.ToInt64(value), 10)
	case "B": // Double
		return strconv.FormatFloat(dbf.ToFloat64(value), 'f', -1, 64)
	case "Y": // Currency
		return strconv.FormatFloat(dbf.ToFloat64(value), 'f', 4, 64)
	default:
		return fmt.Sprintf("%v", value)
	}
}
