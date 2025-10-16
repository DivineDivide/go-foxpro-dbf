# DBF Reader CLI Tool

A command-line tool to read and inspect FoxPro DBF files using the go-foxpro-dbf library.

## Usage

```powershell
go run main.go <DBF_FILE_PATH>
```

## Examples

From the `cmd/dbfreader` directory:

```powershell
# Read a DBF file from testdata
go run main.go ../../testdata/TEST.DBF

# Read any DBF file by providing full path
go run main.go C:\path\to\your\file.dbf
```

## What it shows

- Basic file information (total records, field count, field names)
- Detailed field information (name, type, length, decimals)
- First 10 records (or all if less than 10)
- Properly formatted field values based on their types
- Record navigation demonstration
- JSON export of the first record

## Supported Field Types

The tool displays various DBF field types:
- **C** (Character) - displayed as quoted strings, trimmed
- **N** (Numeric) - displayed as integers (if no decimals) or floats
- **F** (Float) - displayed as floating point numbers  
- **L** (Logical) - displayed as true/false
- **D** (Date) / **T** (DateTime) - displayed in standard time format
- **M** (Memo) - displays memo field content from FPT files
- **I** (Integer) - displayed as integers

## Note

The tool uses Windows-1250 encoding by default, which is common for FoxPro files on Windows platforms.