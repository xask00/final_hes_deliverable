#ifndef DLMS_SHIM_H
#define DLMS_SHIM_H

#include <stdint.h>
#include <time.h>

#ifdef __cplusplus
extern "C" {
#endif

// Configuration structure for energy meter connection
typedef struct {
    char* meter_ip;
    int meter_port;
    int connection_timeout;
    char* auth_password;
    char* system_title;
    char* block_cipher_key;
    char* authentication_key;
    int client_address;
    int server_address;
    int attribute_index;
    int max_entries;
    
    // Debug settings
    int debug_packets;  // Enable raw packet debugging
    
    // Connection state (private - managed by shim)
    void* connection;  // Pointer to connection struct
    int is_connected;
} meter_t;

// Result structure for profile data
typedef struct {
    char* error_message;
    int error_code;
    int num_rows;
    int num_columns;
    char** column_names;
    char** data; // Flattened array: data[row * num_columns + col]
} dlms_result_t;

// Profile Generic object structure (mirrors gxProfileGeneric)
typedef struct {
    void* internal_pg;  // Pointer to actual gxProfileGeneric
    char* logical_name;
    int num_capture_objects;
    char** capture_object_names;
    int buffer_size;
    int entries_in_use;
    int profile_entries;
    int sort_method;
    void* sort_object;
    int capture_period;
} profile_generic_t;

// Initialize a meter configuration with default values
meter_t* meter_create(void);

// Free a meter configuration
void meter_free(meter_t* meter);

// Set configuration parameters (copies the strings)
int meter_set_ip(meter_t* meter, const char* ip);
int meter_set_port(meter_t* meter, int port);
int meter_set_connection_timeout(meter_t* meter, int timeout);
int meter_set_auth_password(meter_t* meter, const char* password);
int meter_set_system_title(meter_t* meter, const char* title);
int meter_set_block_cipher_key(meter_t* meter, const char* key);
int meter_set_authentication_key(meter_t* meter, const char* key);
int meter_set_client_address(meter_t* meter, int address);
int meter_set_server_address(meter_t* meter, int address);
int meter_set_attribute_index(meter_t* meter, int index);
int meter_set_max_entries(meter_t* meter, int max_entries);

// Debug functions
int meter_set_debug_packets(meter_t* meter, int enable);

// Connection management
int meter_connect(meter_t* meter);
int meter_disconnect(meter_t* meter);
int meter_is_connected(meter_t* meter);

// Main function to read profile generic data from DLMS meter (requires connection)
dlms_result_t* meter_read_profile_generic(meter_t* meter, const char* obis_code);

// New separated functions for profile generic operations
profile_generic_t* meter_read_profile_generic_object(meter_t* meter, const char* obis_code);
dlms_result_t* profile_generic_read_rows(meter_t* meter, profile_generic_t* pg, int index, int count);
void profile_generic_free(profile_generic_t* pg);

// Free the result structure
void dlms_result_free(dlms_result_t* result);

// Get a specific data value as string
const char* dlms_result_get_data(dlms_result_t* result, int row, int col);

// Get column name
const char* dlms_result_get_column_name(dlms_result_t* result, int col);

// Write operations
int meter_write_obis_int8(meter_t* meter, const char* obis_code, int8_t value, int object_type, int attribute_index);
int meter_write_obis_int16(meter_t* meter, const char* obis_code, int16_t value, int object_type, int attribute_index);
int meter_write_obis_int32(meter_t* meter, const char* obis_code, int32_t value, int object_type, int attribute_index);
int meter_write_obis_uint8(meter_t* meter, const char* obis_code, uint8_t value, int object_type, int attribute_index);
int meter_write_obis_uint16(meter_t* meter, const char* obis_code, uint16_t value, int object_type, int attribute_index);
int meter_write_obis_uint32(meter_t* meter, const char* obis_code, uint32_t value, int object_type, int attribute_index);
int meter_write_obis_float32(meter_t* meter, const char* obis_code, float value, int object_type, int attribute_index);
int meter_write_obis_float64(meter_t* meter, const char* obis_code, double value, int object_type, int attribute_index);
int meter_write_obis_string(meter_t* meter, const char* obis_code, const char* value, int object_type, int attribute_index);
int meter_write_obis_datetime(meter_t* meter, const char* obis_code, time_t timestamp, int object_type, int attribute_index);
int meter_write_obis_boolean(meter_t* meter, const char* obis_code, unsigned char value, int object_type, int attribute_index);
int meter_write_obis_octet_string(meter_t* meter, const char* obis_code, const unsigned char* data, int length, int object_type, int attribute_index);

/*******************************************************************************
 * Method Invocation Functions
 ******************************************************************************/
int meter_call_method_no_params(meter_t* meter, const char* obis_code, int object_type, int method_index);
int meter_call_method_with_data(meter_t* meter, const char* obis_code, int object_type, int method_index, const unsigned char* data, int data_length);
int meter_call_set_time(meter_t* meter, time_t timestamp);
/******************************************************************************/

// Association View structures and functions
typedef struct {
    char* logical_name;
    uint16_t object_type;
    unsigned char version;
    char** attribute_access_modes;
    char** method_access_modes;
    int num_attributes;
    int num_methods;
} association_object_t;

typedef struct {
    association_object_t* objects;
    int num_objects;
    char* error_message;
    int error_code;
} association_view_t;

// Association view functions
association_view_t* meter_get_association_view(meter_t* meter);
void association_view_free(association_view_t* view);

#ifdef __cplusplus
}
#endif

#endif // DLMS_SHIM_H 