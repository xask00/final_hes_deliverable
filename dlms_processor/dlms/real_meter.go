package dlms

/*
#cgo CFLAGS: -I./include -I./helpers/include
#cgo LDFLAGS: -L./lib dlms/helpers/connection.o dlms/helpers/communication.o -lgurux_dlms_c -lm -lpthread
#include "dlms_shim.h"
#include <stdlib.h>
#include <stdint.h>
#include <time.h>
*/
import "C"
import (
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

func testRealMeter() {
	// Configure the client with the same settings as the C program
	meter := RealMeter{
		MeterIP: "2401:4900:833f:2688:0000:0000:0000:0002",
		// MeterIP:           "2401:4900:833f:2688:0000:0000:0000:0002",
		MeterPort:         4059,
		ConnectionTimeout: 5000,
		// AuthPassword:      "0000000000000000",
		// SystemTitle:       "6162636465666768", // "abcdefgh" in hex

		AuthPassword: "0000000000000000",
		SystemTitle:  "6162636465666768",

		// BlockCipherKey:    "49423031494230324942303349423034", // "IB01IB02IB03IB04" in hex
		// AuthenticationKey: "49423031494230324942303349423034", // "IB01IB02IB03IB04" in hex

		BlockCipherKey:    "49423031494230324942303349423034", // "IB01IB02IB03IB04" in hex
		AuthenticationKey: "49423031494230324942303349423034", // "IB01IB02IB03IB04" in hex

		ClientAddress:  48,
		ServerAddress:  1,
		AttributeIndex: 3,
		MaxEntries:     10,
	}
	fmt.Println(meter)

	client := NewMeterClient()
	if client == nil {
		fmt.Println("Failed to create meter client")
		return
	}

	if err := client.Configure(&meter); err != nil {
		fmt.Printf("Failed to configure meter: %v\n", err)
		return
	}

}

// Meter represents the configuration for connecting to a DLMS energy meter
type RealMeter struct {
	MeterIP           string
	MeterPort         int
	ConnectionTimeout int
	AuthPassword      string
	SystemTitle       string
	BlockCipherKey    string
	AuthenticationKey string
	ClientAddress     int
	ServerAddress     int
	AttributeIndex    int
	MaxEntries        int

	client *MeterClient
}

func NewRealMeter(meter RealMeter) (*RealMeter, error) {
	return &meter, nil
}

func (m *RealMeter) SetClock(clock string) error {
	return nil
}

func (m *RealMeter) ExecuteFunction(function string, params []string) (string, error) {
	return "123", nil
}

func (m *RealMeter) FOTA() error {
	return nil
}

func (m *RealMeter) Connect() error {
	m.client = NewMeterClient()
	if m.client == nil {
		fmt.Println("Failed to create meter client")
		return fmt.Errorf("failed to create meter client")
	}

	if err := m.client.Configure(m); err != nil {
		fmt.Printf("Failed to configure meter: %v\n", err)
		return fmt.Errorf("failed to configure meter: %w", err)
	}

	if m.client != nil {
		fmt.Println("Connect() method - Client is not nil")
	}

	return nil
}

type BlockLoadProfile struct {
	DateTime             string  `obis:"0.0.1.0.0.255" type:"string" json:"date_time"`                  // Real Time Clock (corrected OBIS)
	AverageVoltage       float64 `obis:"1.0.12.27.0.255" type:"float64" json:"average_voltage"`         // Average Voltage
	BlockEnergyWhImport  float64 `obis:"1.0.1.29.0.255" type:"float64" json:"block_energy_wh_import"`   // Block energy Wh-(import)
	BlockEnergyVAhImport float64 `obis:"1.0.9.29.0.255" type:"float64" json:"block_energy_vah_import"`  // Block energy VAh-(import)
	BlockEnergyWhExport  float64 `obis:"1.0.2.29.0.255" type:"float64" json:"block_energy_wh_export"`   // Block energy Wh-export
	BlockEnergyVAhExport float64 `obis:"1.0.10.29.0.255" type:"float64" json:"block_energy_vah_export"` // Block energy VAh-export
	AverageCurrent       float64 `obis:"1.0.11.27.0.255" type:"float64" json:"average_current"`         // Average Current
	MeterHealthIndicator uint8   `obis:"0.0.96.10.1.255" type:"uint8" json:"meter_health_indicator"`    // Meter Health Indicator
}

func (m *RealMeter) GetOBIS(obis string) (string, error) {
	if m.client == nil {
		slog.Error("client not initialized")
		return "", fmt.Errorf("client not initialized")
	}

	err := m.client.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect to meter: %w", err)
	}

	return obis, nil
}

func (m *RealMeter) GetBlockLoadProfile() (*BlockLoadProfile, error) {
	if m.client == nil {
		slog.Error("client not initialized")
		return nil, fmt.Errorf("client not initialized")
	}

	err := m.client.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to meter: %w", err)
	}

	results, err := ReadProfileDataTyped[BlockLoadProfile](m.client, "1.0.99.1.0.255", 1, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile data: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	slog.Info("block load profile results", "results", results[0])

	return &results[0], nil
}

// -------------------- DLMS Meter functions ----------------------

// MeterClient provides a high-level interface for connecting to DLMS meters
type MeterClient struct {
	meter *C.meter_t
}

// NewMeterClient creates a new DLMS meter client with default configuration
func NewMeterClient() *MeterClient {
	meter := C.meter_create()
	if meter == nil {
		return nil
	}

	client := &MeterClient{meter: meter}

	// Set finalizer to ensure cleanup
	runtime.SetFinalizer(client, (*MeterClient).cleanup)

	return client
}

// cleanup ensures the C resources are freed
func (c *MeterClient) cleanup() {
	if c.meter != nil {
		C.meter_free(c.meter)
		c.meter = nil
	}
}

// Close explicitly frees the client resources
func (c *MeterClient) Close() {
	runtime.SetFinalizer(c, nil)
	c.cleanup()
}

// Configure sets multiple configuration parameters at once
func (c *MeterClient) Configure(meter *RealMeter) error {
	if err := c.SetMeterIP(meter.MeterIP); err != nil {
		return fmt.Errorf("setting meter IP: %w", err)
	}

	if meter.MeterPort > 0 {
		if err := c.SetMeterPort(meter.MeterPort); err != nil {
			return fmt.Errorf("setting meter port: %w", err)
		}
	}

	if meter.ConnectionTimeout > 0 {
		if err := c.SetConnectionTimeout(meter.ConnectionTimeout); err != nil {
			return fmt.Errorf("setting connection timeout: %w", err)
		}
	}

	if meter.AuthPassword != "" {
		if err := c.SetAuthPassword(meter.AuthPassword); err != nil {
			return fmt.Errorf("setting auth password: %w", err)
		}
	}

	if meter.SystemTitle != "" {
		if err := c.SetSystemTitle(meter.SystemTitle); err != nil {
			return fmt.Errorf("setting system title: %w", err)
		}
	}

	if meter.BlockCipherKey != "" {
		if err := c.SetBlockCipherKey(meter.BlockCipherKey); err != nil {
			return fmt.Errorf("setting block cipher key: %w", err)
		}
	}

	if meter.AuthenticationKey != "" {
		if err := c.SetAuthenticationKey(meter.AuthenticationKey); err != nil {
			return fmt.Errorf("setting authentication key: %w", err)
		}
	}

	if meter.ClientAddress > 0 {
		if err := c.SetClientAddress(meter.ClientAddress); err != nil {
			return fmt.Errorf("setting client address: %w", err)
		}
	}

	if meter.ServerAddress > 0 {
		if err := c.SetServerAddress(meter.ServerAddress); err != nil {
			return fmt.Errorf("setting server address: %w", err)
		}
	}

	if meter.AttributeIndex > 0 {
		if err := c.SetAttributeIndex(meter.AttributeIndex); err != nil {
			return fmt.Errorf("setting attribute index: %w", err)
		}
	}

	if meter.MaxEntries > 0 {
		if err := c.SetMaxEntries(meter.MaxEntries); err != nil {
			return fmt.Errorf("setting max entries: %w", err)
		}
	}

	return nil
}

// SetMeterIP sets the meter IP address
func (c *MeterClient) SetMeterIP(ip string) error {
	if c.meter == nil {
		return fmt.Errorf("client not initialized")
	}

	cIP := C.CString(ip)
	defer C.free(unsafe.Pointer(cIP))

	ret := C.meter_set_ip(c.meter, cIP)
	if ret != 0 {
		return fmt.Errorf("failed to set meter IP: %d", ret)
	}

	return nil
}

// SetMeterPort sets the meter port
func (c *MeterClient) SetMeterPort(port int) error {
	if c.meter == nil {
		return fmt.Errorf("client not initialized")
	}

	ret := C.meter_set_port(c.meter, C.int(port))
	if ret != 0 {
		return fmt.Errorf("failed to set meter port: %d", ret)
	}

	return nil
}

// SetConnectionTimeout sets the connection timeout in milliseconds
func (c *MeterClient) SetConnectionTimeout(timeout int) error {
	if c.meter == nil {
		return fmt.Errorf("client not initialized")
	}

	ret := C.meter_set_connection_timeout(c.meter, C.int(timeout))
	if ret != 0 {
		return fmt.Errorf("failed to set connection timeout: %d", ret)
	}

	return nil
}

// SetAuthPassword sets the authentication password
func (c *MeterClient) SetAuthPassword(password string) error {
	if c.meter == nil {
		return fmt.Errorf("client not initialized")
	}

	cPassword := C.CString(password)
	defer C.free(unsafe.Pointer(cPassword))

	ret := C.meter_set_auth_password(c.meter, cPassword)
	if ret != 0 {
		return fmt.Errorf("failed to set auth password: %d", ret)
	}

	return nil
}

// SetSystemTitle sets the system title (8 bytes as hex string)
func (c *MeterClient) SetSystemTitle(title string) error {
	if c.meter == nil {
		return fmt.Errorf("client not initialized")
	}

	if len(title) != 16 {
		return fmt.Errorf("system title must be exactly 16 hex characters (8 bytes)")
	}

	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	ret := C.meter_set_system_title(c.meter, cTitle)
	if ret != 0 {
		return fmt.Errorf("failed to set system title: %d", ret)
	}

	return nil
}

// SetBlockCipherKey sets the block cipher key (16 bytes as hex string)
func (c *MeterClient) SetBlockCipherKey(key string) error {
	if c.meter == nil {
		return fmt.Errorf("client not initialized")
	}

	if len(key) != 32 {
		return fmt.Errorf("block cipher key must be exactly 32 hex characters (16 bytes)")
	}

	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	ret := C.meter_set_block_cipher_key(c.meter, cKey)
	if ret != 0 {
		return fmt.Errorf("failed to set block cipher key: %d", ret)
	}

	return nil
}

// SetAuthenticationKey sets the authentication key (16 bytes as hex string)
func (c *MeterClient) SetAuthenticationKey(key string) error {
	if c.meter == nil {
		return fmt.Errorf("client not initialized")
	}

	if len(key) != 32 {
		return fmt.Errorf("authentication key must be exactly 32 hex characters (16 bytes)")
	}

	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	ret := C.meter_set_authentication_key(c.meter, cKey)
	if ret != 0 {
		return fmt.Errorf("failed to set authentication key: %d", ret)
	}

	return nil
}

// SetClientAddress sets the client address
func (c *MeterClient) SetClientAddress(address int) error {
	if c.meter == nil {
		return fmt.Errorf("client not initialized")
	}

	ret := C.meter_set_client_address(c.meter, C.int(address))
	if ret != 0 {
		return fmt.Errorf("failed to set client address: %d", ret)
	}

	return nil
}

// SetServerAddress sets the server address
func (c *MeterClient) SetServerAddress(address int) error {
	if c.meter == nil {
		return fmt.Errorf("client not initialized")
	}

	ret := C.meter_set_server_address(c.meter, C.int(address))
	if ret != 0 {
		return fmt.Errorf("failed to set server address: %d", ret)
	}

	return nil
}

// SetAttributeIndex sets the attribute index
func (c *MeterClient) SetAttributeIndex(index int) error {
	if c.meter == nil {
		return fmt.Errorf("client not initialized")
	}

	ret := C.meter_set_attribute_index(c.meter, C.int(index))
	if ret != 0 {
		return fmt.Errorf("failed to set attribute index: %d", ret)
	}

	return nil
}

// SetMaxEntries sets the maximum number of entries to read
func (c *MeterClient) SetMaxEntries(maxEntries int) error {
	if c.meter == nil {
		return fmt.Errorf("client not initialized")
	}

	ret := C.meter_set_max_entries(c.meter, C.int(maxEntries))
	if ret != 0 {
		return fmt.Errorf("failed to set max entries: %d", ret)
	}

	return nil
}

// Connect establishes a connection to the DLMS meter
func (c *MeterClient) Connect() error {
	if c.meter == nil {
		return fmt.Errorf("client not initialized")
	}

	ret := C.meter_connect(c.meter)
	if ret != 0 {
		return fmt.Errorf("failed to connect to meter: error code %d", ret)
	}

	return nil
}

func ReadProfileDataTyped[T any](c *MeterClient, obisCode string, index, count int) ([]T, error) {
	var zero T
	structType := reflect.TypeOf(zero)

	genericResults, err := c.ReadProfileData(obisCode, index, count, structType)
	if err != nil {
		return nil, err
	}

	// Convert generic results to typed slice
	results := make([]T, len(genericResults))
	for i, genericResult := range genericResults {
		results[i] = genericResult.(T)
	}

	return results, nil
}

// ReadProfileData is a generic function that reads profile data and maps it to any struct type with OBIS tags
func (c *MeterClient) ReadProfileData(obisCode string, index, count int, structType reflect.Type) ([]interface{}, error) {
	// First get the raw data using the existing method
	result, err := c.ProfileGenericReadRows(obisCode, index, count)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile data: %w", err)
	}

	if result.ErrorCode != 0 {
		return nil, fmt.Errorf("DLMS error %d: %s", result.ErrorCode, result.ErrorMessage)
	}

	// slog trace level print result
	slog.Info("result", "result", result)

	// Use the generic mapping function
	return mapDLMSDataToStruct(result, structType)
}

// DLMSResult represents the result of reading profile data from a DLMS meter
type DLMSResult struct {
	ErrorCode    int
	ErrorMessage string
	NumRows      int
	NumColumns   int
	ColumnNames  []string
	Data         [][]string
}

// mapDLMSDataToStruct uses reflection to map DLMS data to any struct with OBIS tags
func mapDLMSDataToStruct(result *DLMSResult, structType reflect.Type) ([]interface{}, error) {
	if result.NumRows == 0 {
		return []interface{}{}, nil
	}

	// Create column mapping from actual column names
	columnMap := make(map[string]int)
	for i, columnName := range result.ColumnNames {
		columnMap[columnName] = i
	}

	// Create slice to hold the results
	results := make([]interface{}, result.NumRows)

	// Process each row
	for rowIdx := 0; rowIdx < result.NumRows; rowIdx++ {
		// Create new instance of the struct
		structValue := reflect.New(structType).Elem()
		row := result.Data[rowIdx]

		// Iterate through struct fields
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			fieldValue := structValue.Field(i)

			// Skip unexported fields
			if !fieldValue.CanSet() {
				continue
			}

			// Get OBIS code and type from struct tags
			obisCode := field.Tag.Get("obis")
			dataType := field.Tag.Get("type")

			if obisCode == "" || dataType == "" {
				continue // Skip fields without proper tags
			}

			// Find column index for this OBIS code
			colIdx := -1
			exists := false

			// First try exact OBIS code match
			if idx, found := columnMap[obisCode]; found {
				colIdx = idx
				exists = true
			} else {
				// Try to find by field name as fallback
				for j, columnName := range result.ColumnNames {
					if strings.Contains(strings.ToLower(columnName), strings.ToLower(field.Name)) {
						colIdx = j
						exists = true
						break
					}
				}

				// If still not found, try partial OBIS matching (in case format is different)
				if !exists {
					for j, columnName := range result.ColumnNames {
						// Try to see if the column name contains the OBIS code or vice versa
						if strings.Contains(columnName, obisCode) || strings.Contains(obisCode, columnName) {
							colIdx = j
							exists = true
							break
						}
					}
				}

				// Special case for datetime/clock fields - try common patterns
				if !exists && (field.Name == "DateTime" || strings.Contains(strings.ToLower(field.Name), "time") || strings.Contains(strings.ToLower(field.Name), "date")) {
					for j, columnName := range result.ColumnNames {
						lowerCol := strings.ToLower(columnName)
						if strings.Contains(lowerCol, "time") || strings.Contains(lowerCol, "date") || strings.Contains(lowerCol, "clock") || strings.Contains(lowerCol, "rtc") {
							colIdx = j
							exists = true
							break
						}
					}
				}
			}

			if !exists || colIdx >= len(row) || colIdx < 0 {
				continue // Skip if column not found
			}

			cellValue := row[colIdx]

			// Parse the value according to the specified type
			parsedValue, err := parseValueByType(cellValue, dataType)
			if err != nil {
				continue // Skip if parsing fails
			}

			// Set the field value using reflection
			switch dataType {
			case "string":
				if fieldValue.Kind() == reflect.String {
					fieldValue.SetString(parsedValue.(string))
				}
			case "float64":
				if fieldValue.Kind() == reflect.Float64 {
					fieldValue.SetFloat(parsedValue.(float64))
				}
			case "uint8":
				if fieldValue.Kind() == reflect.Uint8 {
					fieldValue.SetUint(uint64(parsedValue.(uint8)))
				}
			case "int":
				if fieldValue.Kind() == reflect.Int {
					fieldValue.SetInt(int64(parsedValue.(int)))
				}
			case "uint16":
				if fieldValue.Kind() == reflect.Uint16 {
					fieldValue.SetUint(uint64(parsedValue.(uint16)))
				}
			case "uint32":
				if fieldValue.Kind() == reflect.Uint32 {
					fieldValue.SetUint(uint64(parsedValue.(uint32)))
				}
			}
		}

		results[rowIdx] = structValue.Interface()
	}

	return results, nil
}

// parseValueByType parses a string value to the specified type
func parseValueByType(value string, targetType string) (interface{}, error) {
	switch targetType {
	case "string":
		return value, nil
	case "float64":
		return parseFloat64(value)
	case "uint8":
		return parseUint8(value)
	case "int":
		if val, err := strconv.Atoi(value); err == nil {
			return val, nil
		}
		return 0, fmt.Errorf("failed to parse int from '%s'", value)
	case "uint16":
		if val, err := strconv.ParseUint(value, 10, 16); err == nil {
			return uint16(val), nil
		}
		return uint16(0), fmt.Errorf("failed to parse uint16 from '%s'", value)
	case "uint32":
		if val, err := strconv.ParseUint(value, 10, 32); err == nil {
			return uint32(val), nil
		}
		return uint32(0), fmt.Errorf("failed to parse uint32 from '%s'", value)
	default:
		return nil, fmt.Errorf("unsupported type: %s", targetType)
	}
}

// parseFloat64 safely parses a string to float64, handling common DLMS data formats
func parseFloat64(s string) (float64, error) {
	if s == "" || s == "[error]" || s == "null" {
		return 0.0, fmt.Errorf("invalid value: %s", s)
	}

	// Remove any whitespace
	s = strings.TrimSpace(s)

	// Try to parse as float64
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, fmt.Errorf("failed to parse float64 from '%s': %w", s, err)
	}

	return val, nil
}

// parseUint8 safely parses a string to uint8, handling common DLMS data formats
func parseUint8(s string) (uint8, error) {
	if s == "" || s == "[error]" || s == "null" {
		return 0, fmt.Errorf("invalid value: %s", s)
	}

	// Remove any whitespace
	s = strings.TrimSpace(s)

	// Try to parse as uint64 first, then convert to uint8
	val, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uint8 from '%s': %w", s, err)
	}

	return uint8(val), nil
}

// ProfileGenericReadRows reads rows from a profile generic object using OBIS code
func (c *MeterClient) ProfileGenericReadRows(obisCode string, index, count int) (*DLMSResult, error) {
	if c.meter == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	if obisCode == "" {
		return nil, fmt.Errorf("OBIS code cannot be empty")
	}

	// Use the OBIS code to read the profile data directly
	cObisCode := C.CString(obisCode)
	defer C.free(unsafe.Pointer(cObisCode))

	// Call the C function to read profile generic data
	cResult := C.meter_read_profile_generic(c.meter, cObisCode)
	if cResult == nil {
		return nil, fmt.Errorf("failed to read profile generic rows: C function returned NULL")
	}

	// Ensure cleanup of C result
	defer C.dlms_result_free(cResult)

	// Convert C result to Go result
	result := &DLMSResult{
		ErrorCode:    int(cResult.error_code),
		ErrorMessage: C.GoString(cResult.error_message),
		NumRows:      int(cResult.num_rows),
		NumColumns:   int(cResult.num_columns),
	}

	// Check for errors
	if result.ErrorCode != 0 {
		return result, fmt.Errorf("DLMS error %d: %s", result.ErrorCode, result.ErrorMessage)
	}

	// Extract column names
	if result.NumColumns > 0 {
		result.ColumnNames = make([]string, result.NumColumns)
		for i := 0; i < result.NumColumns; i++ {
			columnName := C.dlms_result_get_column_name(cResult, C.int(i))
			if columnName != nil {
				result.ColumnNames[i] = C.GoString(columnName)
			} else {
				result.ColumnNames[i] = fmt.Sprintf("Column_%d", i)
			}
		}
	}

	// Extract data
	if result.NumRows > 0 && result.NumColumns > 0 {
		result.Data = make([][]string, result.NumRows)
		for row := 0; row < result.NumRows; row++ {
			result.Data[row] = make([]string, result.NumColumns)
			for col := 0; col < result.NumColumns; col++ {
				cellData := C.dlms_result_get_data(cResult, C.int(row), C.int(col))
				if cellData != nil {
					result.Data[row][col] = C.GoString(cellData)
				} else {
					result.Data[row][col] = "[error]"
				}
			}
		}
	}

	return result, nil
}
