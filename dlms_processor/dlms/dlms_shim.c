#include "dlms_shim.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include <ctype.h>

#if defined(_WIN32) || defined(_WIN64)
#include <conio.h>
#include <process.h>
#include <time.h>
#if _MSC_VER > 1000
#include <crtdbg.h>
#endif
#define strcasecmp _strcmpi
#else
#include <unistd.h>
#include <strings.h>
#include <sys/time.h>
#include <time.h>
#endif

#include "./helpers/include/communication.h"
#include "./helpers/include/connection.h"
#include "./include/gxserializer.h"

// Default values
#define DEFAULT_PORT 4059
#define DEFAULT_TIMEOUT 5000
#define DEFAULT_CLIENT_ADDRESS 48
#define DEFAULT_SERVER_ADDRESS 1
#define DEFAULT_ATTRIBUTE_INDEX 3
#define DEFAULT_MAX_ENTRIES 10

// Authentication and security defaults
#define DEFAULT_AUTH_TYPE DLMS_AUTHENTICATION_HIGH_GMAC
#define DEFAULT_SECURITY_LEVEL DLMS_SECURITY_AUTHENTICATION_ENCRYPTION
#define DEFAULT_INTERFACE_TYPE DLMS_INTERFACE_TYPE_WRAPPER

// Stub functions required by the framework
void svr_preGet(dlmsSettings* settings, gxValueEventCollection* args) {}
void svr_postGet(dlmsSettings* settings, gxValueEventCollection* args) {}
void svr_preRead(dlmsSettings* settings, gxValueEventCollection* args) {}
void svr_preWrite(dlmsSettings* settings, gxValueEventCollection* args) {}
void svr_preAction(dlmsSettings* settings, gxValueEventCollection* args) {}
void svr_postRead(dlmsSettings* settings, gxValueEventCollection* args) {}
void svr_postWrite(dlmsSettings* settings, gxValueEventCollection* args) {}
void svr_postAction(dlmsSettings* settings, gxValueEventCollection* args) {}
unsigned char svr_isTarget(dlmsSettings* settings, unsigned long serverAddress, unsigned long clientAddress) { return 0; }
int svr_connected(dlmsServerSettings* settings) { return 0; }
DLMS_ACCESS_MODE svr_getAttributeAccess(dlmsSettings* settings, gxObject* obj, unsigned char index) { return DLMS_ACCESS_MODE_READ_WRITE; }
DLMS_METHOD_ACCESS_MODE svr_getMethodAccess(dlmsSettings* settings, gxObject* obj, unsigned char index) { return DLMS_METHOD_ACCESS_MODE_ACCESS; }
void svr_trace(const char* str, const char* data) {}

// Helper function to safely copy strings
static char* safe_strdup(const char* str) {
    if (!str) return NULL;
    size_t len = strlen(str);
    char* copy = malloc(len + 1);
    if (copy) {
        strcpy(copy, str);
    }
    return copy;
}

// Helper function to convert hex string to bytes
static int hex_to_bytes(const char* hex_str, unsigned char* bytes, int max_bytes) {
    if (!hex_str || !bytes) return -1;
    
    int len = strlen(hex_str);
    if (len % 2 != 0 || len / 2 > max_bytes) return -1;
    
    for (int i = 0, j = 0; i < len; i += 2, j++) {
        char hex[3] = {hex_str[i], hex_str[i+1], 0};
        int val;
        if (sscanf(hex, "%x", &val) != 1) return -1;
        bytes[j] = (unsigned char)val;
    }
    
    return len / 2;
}

// Helper function to print raw packet data in hex format
static void print_packet_hex(const char* direction, const unsigned char* data, int length) {
    printf("\n=== %s PACKET ===\n", direction);
    printf("Length: %d bytes\n", length);
    printf("Raw data:\n");
    
    for (int i = 0; i < length; i++) {
        if (i % 16 == 0) {
            printf("%04X: ", i);
        }
        printf("%02X ", data[i]);
        if ((i + 1) % 16 == 0 || i == length - 1) {
            // Print ASCII representation
            int ascii_start = (i / 16) * 16;
            int ascii_count = (i % 16) + 1;
            if (i == length - 1 && i % 16 != 15) {
                // Pad spaces for last incomplete line
                for (int j = ascii_count; j < 16; j++) {
                    printf("   ");
                }
            }
            printf(" |");
            for (int j = 0; j < ascii_count; j++) {
                unsigned char c = data[ascii_start + j];
                printf("%c", (c >= 32 && c <= 126) ? c : '.');
            }
            printf("|\n");
        }
    }
    printf("=== END PACKET ===\n\n");
}

// Helper function to convert variant to string
static char* variant_to_string(dlmsVARIANT* value) {
    if (!value) return safe_strdup("[NULL]");
    
    char buffer[1024] = {0};
    
    switch (value->vt) {
        case DLMS_DATA_TYPE_BOOLEAN:
            snprintf(buffer, sizeof(buffer), "%s", value->boolVal ? "true" : "false");
            break;
        case DLMS_DATA_TYPE_INT8:
            snprintf(buffer, sizeof(buffer), "%d", (int)value->cVal);
            break;
        case DLMS_DATA_TYPE_INT16:
            snprintf(buffer, sizeof(buffer), "%d", value->iVal);
            break;
        case DLMS_DATA_TYPE_INT32:
            snprintf(buffer, sizeof(buffer), "%d", value->lVal);
            break;
        case DLMS_DATA_TYPE_INT64:
            snprintf(buffer, sizeof(buffer), "%lld", (long long)value->llVal);
            break;
        case DLMS_DATA_TYPE_UINT8:
            snprintf(buffer, sizeof(buffer), "%u", (unsigned int)value->bVal);
            break;
        case DLMS_DATA_TYPE_UINT16:
            snprintf(buffer, sizeof(buffer), "%u", value->uiVal);
            break;
        case DLMS_DATA_TYPE_UINT32:
            snprintf(buffer, sizeof(buffer), "%u", value->ulVal);
            break;
        case DLMS_DATA_TYPE_UINT64:
            snprintf(buffer, sizeof(buffer), "%llu", (unsigned long long)value->ullVal);
            break;
        case DLMS_DATA_TYPE_ENUM:
            snprintf(buffer, sizeof(buffer), "Enum:%u", value->uiVal);
            break;
#ifndef DLMS_IGNORE_FLOAT32
        case DLMS_DATA_TYPE_FLOAT32:
            snprintf(buffer, sizeof(buffer), "%f", value->fltVal);
            break;
#endif
#ifndef DLMS_IGNORE_FLOAT64
        case DLMS_DATA_TYPE_FLOAT64:
            snprintf(buffer, sizeof(buffer), "%lf", value->dblVal);
            break;
#endif
        case DLMS_DATA_TYPE_STRING:
        case DLMS_DATA_TYPE_STRING_UTF8: {
            gxByteBuffer* strBytes = (value->vt == DLMS_DATA_TYPE_STRING) ? value->strVal : value->strUtfVal;
            if (strBytes && strBytes->size > 0) {
                int copy_len = strBytes->size < sizeof(buffer) - 1 ? strBytes->size : sizeof(buffer) - 1;
                memcpy(buffer, strBytes->data, copy_len);
                buffer[copy_len] = '\0';
            } else {
                strcpy(buffer, "[empty string]");
            }
            break;
        }
        case DLMS_DATA_TYPE_OCTET_STRING:
        case DLMS_DATA_TYPE_BIT_STRING:
        case DLMS_DATA_TYPE_BINARY_CODED_DESIMAL: {
            gxByteBuffer* byteData = (value->vt == DLMS_DATA_TYPE_BIT_STRING) ? 
                (gxByteBuffer*)value->bitArr : value->byteArr;
            if (byteData && byteData->size > 0) {
                snprintf(buffer, sizeof(buffer), "Hex:");
                int pos = 4;
                for (uint16_t i = 0; i < byteData->size && pos < sizeof(buffer) - 3; ++i) {
                    snprintf(buffer + pos, sizeof(buffer) - pos, "%02X", byteData->data[i]);
                    pos += 2;
                }
            } else {
                strcpy(buffer, "[empty bytes]");
            }
            break;
        }
        case DLMS_DATA_TYPE_DATETIME:
        case DLMS_DATA_TYPE_DATE:
        case DLMS_DATA_TYPE_TIME: {
            if (value->dateTime) {
                time_toString2(value->dateTime, buffer, sizeof(buffer));
            } else {
                strcpy(buffer, "[NULL datetime]");
            }
            break;
        }
        default:
            snprintf(buffer, sizeof(buffer), "[Type %d not handled]", value->vt);
            break;
    }
    
    return safe_strdup(buffer);
}

meter_t* meter_create(void) {
    meter_t* meter = calloc(1, sizeof(meter_t));
    if (!meter) return NULL;
    
    // Set defaults
    meter->meter_port = DEFAULT_PORT;
    meter->connection_timeout = DEFAULT_TIMEOUT;
    meter->client_address = DEFAULT_CLIENT_ADDRESS;
    meter->server_address = DEFAULT_SERVER_ADDRESS;
    meter->attribute_index = DEFAULT_ATTRIBUTE_INDEX;
    meter->max_entries = DEFAULT_MAX_ENTRIES;
    
    // Initialize debug settings
    meter->debug_packets = 0;  // Debug off by default
    
    // Initialize connection state
    meter->connection = NULL;
    meter->is_connected = 0;
    
    return meter;
}

void meter_free(meter_t* meter) {
    if (!meter) return;
    
    // Disconnect if connected
    if (meter->is_connected) {
        meter_disconnect(meter);
    }
    
    free(meter->meter_ip);
    free(meter->auth_password);
    free(meter->system_title);
    free(meter->block_cipher_key);
    free(meter->authentication_key);
    free(meter);
}

int meter_set_ip(meter_t* meter, const char* ip) {
    if (!meter) return -1;
    free(meter->meter_ip);
    meter->meter_ip = safe_strdup(ip);
    return meter->meter_ip ? 0 : -1;
}

int meter_set_port(meter_t* meter, int port) {
    if (!meter || port <= 0) return -1;
    meter->meter_port = port;
    return 0;
}

int meter_set_connection_timeout(meter_t* meter, int timeout) {
    if (!meter || timeout <= 0) return -1;
    meter->connection_timeout = timeout;
    return 0;
}

int meter_set_auth_password(meter_t* meter, const char* password) {
    if (!meter) return -1;
    free(meter->auth_password);
    meter->auth_password = safe_strdup(password);
    return meter->auth_password ? 0 : -1;
}

int meter_set_system_title(meter_t* meter, const char* title) {
    if (!meter) return -1;
    free(meter->system_title);
    meter->system_title = safe_strdup(title);
    return meter->system_title ? 0 : -1;
}

int meter_set_block_cipher_key(meter_t* meter, const char* key) {
    if (!meter) return -1;
    free(meter->block_cipher_key);
    meter->block_cipher_key = safe_strdup(key);
    return meter->block_cipher_key ? 0 : -1;
}

int meter_set_authentication_key(meter_t* meter, const char* key) {
    if (!meter) return -1;
    free(meter->authentication_key);
    meter->authentication_key = safe_strdup(key);
    return meter->authentication_key ? 0 : -1;
}

int meter_set_client_address(meter_t* meter, int address) {
    if (!meter) return -1;
    meter->client_address = address;
    return 0;
}

int meter_set_server_address(meter_t* meter, int address) {
    if (!meter) return -1;
    meter->server_address = address;
    return 0;
}

int meter_set_attribute_index(meter_t* meter, int index) {
    if (!meter || index < 0) return -1;
    meter->attribute_index = index;
    return 0;
}

int meter_set_max_entries(meter_t* meter, int max_entries) {
    if (!meter || max_entries <= 0) return -1;
    meter->max_entries = max_entries;
    return 0;
}

int meter_set_debug_packets(meter_t* meter, int enable) {
    if (!meter) return -1;
    meter->debug_packets = enable ? 1 : 0;
    return 0;
}

int meter_connect(meter_t* meter) {
    if (!meter || !meter->meter_ip) {
        return -1; // Invalid meter configuration
    }
    
    if (meter->is_connected) {
        return 0; // Already connected
    }
    
    // Allocate connection structure
    connection* con = malloc(sizeof(connection));
    if (!con) return -2; // Memory allocation failed
    
    int ret;
    
    // Initialize connection
    con_init(con, GX_TRACE_LEVEL_ERROR);
    
    // Initialize client settings
    cl_init(&con->settings, 1, meter->client_address, meter->server_address,
           DEFAULT_AUTH_TYPE, NULL, DEFAULT_INTERFACE_TYPE);
    
    con->settings.cipher.security = DEFAULT_SECURITY_LEVEL;
    
    // Set password
    if (meter->auth_password && strlen(meter->auth_password) > 0) {
        bb_init(&con->settings.password);
        bb_addString(&con->settings.password, meter->auth_password);
    }
    
    // Set system title
    if (meter->system_title && strlen(meter->system_title) == 16) {
        bb_init(&con->settings.cipher.systemTitle);
        unsigned char st[8];
        if (hex_to_bytes(meter->system_title, st, 8) == 8) {
            bb_set(&con->settings.cipher.systemTitle, st, 8);
        }
    }
    
    // Set block cipher key
    if (meter->block_cipher_key && strlen(meter->block_cipher_key) == 32) {
        bb_init(&con->settings.cipher.blockCipherKey);
        unsigned char bck[16];
        if (hex_to_bytes(meter->block_cipher_key, bck, 16) == 16) {
            bb_set(&con->settings.cipher.blockCipherKey, bck, 16);
        }
    }
    
    // Set authentication key
    if (meter->authentication_key && strlen(meter->authentication_key) == 32) {
        bb_init(&con->settings.cipher.authenticationKey);
        unsigned char ak[16];
        if (hex_to_bytes(meter->authentication_key, ak, 16) == 16) {
            bb_set(&con->settings.cipher.authenticationKey, ak, 16);
        }
    }
    
    // Connect to meter
    ret = com_makeConnect(con, meter->meter_ip, meter->meter_port, meter->connection_timeout);
    if (ret != DLMS_ERROR_CODE_OK) {
        com_close(con);
        con_close(con);
        cl_clear(&con->settings);
        free(con);
        return ret;
    }
    
    // Initialize connection
    ret = com_initializeConnection(con);
    if (ret != 0) {
        com_close(con);
        con_close(con);
        cl_clear(&con->settings);
        free(con);
        return ret;
    }
    
    // Store connection and mark as connected
    meter->connection = con;
    meter->is_connected = 1;
    
    return 0; // Success
}

int meter_disconnect(meter_t* meter) {
    if (!meter || !meter->is_connected || !meter->connection) {
        return 0; // Already disconnected or invalid
    }
    
    connection* con = (connection*)meter->connection;
    
    // Close connection
    com_close(con);
    con_close(con);
    cl_clear(&con->settings);
    free(con);
    
    // Reset connection state
    meter->connection = NULL;
    meter->is_connected = 0;
    
    return 0;
}

int meter_is_connected(meter_t* meter) {
    if (!meter) return 0;
    return meter->is_connected;
}

dlms_result_t* meter_read_profile_generic(meter_t* meter, const char* obis_code) {
    if (!meter || !obis_code) {
        dlms_result_t* result = calloc(1, sizeof(dlms_result_t));
        if (result) {
            result->error_code = -1;
            result->error_message = safe_strdup("Invalid meter configuration or OBIS code");
        }
        return result;
    }
    
    if (!meter->is_connected || !meter->connection) {
        dlms_result_t* result = calloc(1, sizeof(dlms_result_t));
        if (result) {
            result->error_code = -2;
            result->error_message = safe_strdup("Meter not connected. Call meter_connect() first.");
        }
        return result;
    }
    
    dlms_result_t* result = calloc(1, sizeof(dlms_result_t));
    if (!result) return NULL;
    
    connection* con = (connection*)meter->connection;
    int ret;
    
    // Create profile generic object
    gxProfileGeneric pg;
    unsigned char ln[6];
    
    memset(&pg, 0, sizeof(gxProfileGeneric));
    hlp_setLogicalName(ln, obis_code);
    cosem_init2((gxObject*)&pg, DLMS_OBJECT_TYPE_PROFILE_GENERIC, ln);
    
    arr_init(&pg.buffer);
    arr_init(&pg.captureObjects);
    
    // Read capture objects
    ret = com_read(con, (gxObject*)&pg, 3);
    if (ret != 0) {
        result->error_code = ret;
        result->error_message = safe_strdup(hlp_getErrorMessage(ret));
        goto cleanup_pg;
    }
    
    // Read profile data
    ret = com_readRowsByEntry(con, &pg, 1, meter->max_entries);
    if (ret != 0) {
        result->error_code = ret;
        result->error_message = safe_strdup(hlp_getErrorMessage(ret));
        goto cleanup_pg;
    }
    
    // Process results
    result->num_rows = pg.buffer.size;
    result->num_columns = pg.captureObjects.size;
    
    if (result->num_columns > 0) {
        // Allocate column names
        result->column_names = calloc(result->num_columns, sizeof(char*));
        if (!result->column_names) {
            result->error_code = -1;
            result->error_message = safe_strdup("Memory allocation failed");
            goto cleanup_pg;
        }
        
        // Extract column names from capture objects
        for (int col = 0; col < result->num_columns; col++) {
            void* value = NULL;
            if (arr_getByIndex(&pg.captureObjects, col, &value) == 0 && value) {
                gxKey* kv = (gxKey*)value;
                gxObject* obj = (gxObject*)kv->key;
                char ln[25];
                hlp_getLogicalNameToString(obj->logicalName, ln);
                result->column_names[col] = safe_strdup(ln);
            } else {
                char temp[32];
                snprintf(temp, sizeof(temp), "Column_%d", col);
                result->column_names[col] = safe_strdup(temp);
            }
        }
    }
    
    if (result->num_rows > 0 && result->num_columns > 0) {
        // Allocate data array
        int total_cells = result->num_rows * result->num_columns;
        result->data = calloc(total_cells, sizeof(char*));
        if (!result->data) {
            result->error_code = -1;
            result->error_message = safe_strdup("Memory allocation failed");
            goto cleanup_pg;
        }
        
        // Extract data
        for (int row = 0; row < result->num_rows; row++) {
            void* rowPtr = NULL;
            if (arr_getByIndex(&pg.buffer, row, &rowPtr) == 0 && rowPtr) {
                gxArray* cellsInRow = (gxArray*)rowPtr;
                
                for (int col = 0; col < result->num_columns && col < cellsInRow->size; col++) {
                    void* cellPtr = NULL;
                    if (arr_getByIndex(cellsInRow, col, &cellPtr) == 0 && cellPtr) {
                        dlmsVARIANT* cellValue = (dlmsVARIANT*)cellPtr;
                        result->data[row * result->num_columns + col] = variant_to_string(cellValue);
                    } else {
                        result->data[row * result->num_columns + col] = safe_strdup("[error]");
                    }
                }
            }
        }
    }
    
    result->error_code = 0;
    result->error_message = safe_strdup("Success");
    
cleanup_pg:
    obj_clear((gxObject*)&pg);
    
    return result;
}

profile_generic_t* meter_read_profile_generic_object(meter_t* meter, const char* obis_code) {
    if (!meter || !obis_code) {
        return NULL;
    }
    
    if (!meter->is_connected || !meter->connection) {
        return NULL;
    }
    
    connection* con = (connection*)meter->connection;
    int ret;
    
    // Allocate and initialize profile generic object
    gxProfileGeneric* pg = malloc(sizeof(gxProfileGeneric));
    if (!pg) return NULL;
    
    unsigned char ln[6];
    memset(pg, 0, sizeof(gxProfileGeneric));
    hlp_setLogicalName(ln, obis_code);
    cosem_init2((gxObject*)pg, DLMS_OBJECT_TYPE_PROFILE_GENERIC, ln);
    
    arr_init(&pg->buffer);
    arr_init(&pg->captureObjects);
    
    // Read capture objects (attribute 3)
    ret = com_read(con, (gxObject*)pg, 3);
    if (ret != 0) {
        obj_clear((gxObject*)pg);
        free(pg);
        return NULL;
    }

    // Read buffer size (attribute 4)
    ret = com_read(con, (gxObject*)pg, 4);
    if (ret != 0) {
        printf("Error reading buffer size: %d\n", ret); 
        obj_clear((gxObject*)pg);
        free(pg);
        return NULL;
    }

    // attribute 8
    ret = com_read(con, (gxObject*)pg, 8);
    if (ret != 0) {
        printf("Error reading buffer size: %d\n", ret);
        obj_clear((gxObject*)pg);
        free(pg);
        return NULL;
    }

    // Read buffer size (attribute 6)
    ret = com_read(con, (gxObject*)pg, 6);
    if (ret != 0) {
        printf("Error reading buffer size: %d\n", ret);
        obj_clear((gxObject*)pg);
        free(pg);
        return NULL;
    }

    // Read capture period (attribute 7)
    ret = com_read(con, (gxObject*)pg, 7);
    if (ret != 0) {
        printf("Error reading capture period: %d\n", ret);
        obj_clear((gxObject*)pg);
        free(pg);
        return NULL;
    }
    // Create wrapper structure
    profile_generic_t* wrapper = calloc(1, sizeof(profile_generic_t));
    if (!wrapper) {
        obj_clear((gxObject*)pg);
        free(pg);
        return NULL;
    }
    
    // Store internal pointer
    wrapper->internal_pg = pg;
    
    // Copy logical name
    char ln_str[25];
    hlp_getLogicalNameToString(pg->base.logicalName, ln_str);
    wrapper->logical_name = safe_strdup(ln_str);
    
    // Extract capture objects information
    wrapper->num_capture_objects = pg->captureObjects.size;
    if (wrapper->num_capture_objects > 0) {
        wrapper->capture_object_names = calloc(wrapper->num_capture_objects, sizeof(char*));
        if (wrapper->capture_object_names) {
            for (int i = 0; i < wrapper->num_capture_objects; i++) {
                void* value = NULL;
                if (arr_getByIndex(&pg->captureObjects, i, &value) == 0 && value) {
                    gxKey* kv = (gxKey*)value;
                    gxObject* obj = (gxObject*)kv->key;
                    char obj_ln[25];
                    hlp_getLogicalNameToString(obj->logicalName, obj_ln);
                    wrapper->capture_object_names[i] = safe_strdup(obj_ln);
                } else {
                    char temp[32];
                    snprintf(temp, sizeof(temp), "Object_%d", i);
                    wrapper->capture_object_names[i] = safe_strdup(temp);
                }
            }
        }
    }
    
    // Set other properties (these would need to be read from other attributes if needed)
    wrapper->buffer_size = pg->buffer.size;
    wrapper->entries_in_use = pg->entriesInUse;
    wrapper->profile_entries = pg->profileEntries;
    wrapper->sort_method = pg->sortMethod;
    wrapper->sort_object = pg->sortObject;
    wrapper->capture_period = pg->capturePeriod;
    
    return wrapper;
}

dlms_result_t* profile_generic_read_rows(meter_t* meter, profile_generic_t* pg_wrapper, int index, int count) {
    if (!meter || !pg_wrapper || !pg_wrapper->internal_pg) {
        dlms_result_t* result = calloc(1, sizeof(dlms_result_t));
        if (result) {
            result->error_code = -1;
            result->error_message = safe_strdup("Invalid parameters");
        }
        return result;
    }
    
    if (!meter->is_connected || !meter->connection) {
        dlms_result_t* result = calloc(1, sizeof(dlms_result_t));
        if (result) {
            result->error_code = -2;
            result->error_message = safe_strdup("Meter not connected");
        }
        return result;
    }
    
    connection* con = (connection*)meter->connection;
    gxProfileGeneric* pg = (gxProfileGeneric*)pg_wrapper->internal_pg;
    int ret;
    
    dlms_result_t* result = calloc(1, sizeof(dlms_result_t));
    if (!result) return NULL;
    
    // Clear any existing buffer data
    arr_clear(&pg->buffer);
    
    // Read profile data rows
    ret = com_readRowsByEntry(con, pg, index, count);
    if (ret != 0) {
        result->error_code = ret;
        result->error_message = safe_strdup(hlp_getErrorMessage(ret));
        return result;
    }
    
    // Process results
    result->num_rows = pg->buffer.size;
    result->num_columns = pg->captureObjects.size;
    
    if (result->num_columns > 0) {
        // Allocate column names
        result->column_names = calloc(result->num_columns, sizeof(char*));
        if (!result->column_names) {
            result->error_code = -1;
            result->error_message = safe_strdup("Memory allocation failed");
            return result;
        }
        
        // Use the capture object names from wrapper
        for (int col = 0; col < result->num_columns; col++) {
            if (col < pg_wrapper->num_capture_objects && pg_wrapper->capture_object_names[col]) {
                result->column_names[col] = safe_strdup(pg_wrapper->capture_object_names[col]);
            } else {
                char temp[32];
                snprintf(temp, sizeof(temp), "Column_%d", col);
                result->column_names[col] = safe_strdup(temp);
            }
        }
    }
    
    if (result->num_rows > 0 && result->num_columns > 0) {
        // Allocate data array
        int total_cells = result->num_rows * result->num_columns;
        result->data = calloc(total_cells, sizeof(char*));
        if (!result->data) {
            result->error_code = -1;
            result->error_message = safe_strdup("Memory allocation failed");
            return result;
        }
        
        // Extract data
        for (int row = 0; row < result->num_rows; row++) {
            void* rowPtr = NULL;
            if (arr_getByIndex(&pg->buffer, row, &rowPtr) == 0 && rowPtr) {
                gxArray* cellsInRow = (gxArray*)rowPtr;
                
                for (int col = 0; col < result->num_columns && col < cellsInRow->size; col++) {
                    void* cellPtr = NULL;
                    if (arr_getByIndex(cellsInRow, col, &cellPtr) == 0 && cellPtr) {
                        dlmsVARIANT* cellValue = (dlmsVARIANT*)cellPtr;
                        result->data[row * result->num_columns + col] = variant_to_string(cellValue);
                    } else {
                        result->data[row * result->num_columns + col] = safe_strdup("[error]");
                    }
                }
            }
        }
    }
    
    result->error_code = 0;
    result->error_message = safe_strdup("Success");
    
    return result;
}

void profile_generic_free(profile_generic_t* pg_wrapper) {
    if (!pg_wrapper) return;
    
    // Free the internal gxProfileGeneric
    if (pg_wrapper->internal_pg) {
        gxProfileGeneric* pg = (gxProfileGeneric*)pg_wrapper->internal_pg;
        obj_clear((gxObject*)pg);
        free(pg);
    }
    
    // Free wrapper fields
    free(pg_wrapper->logical_name);
    
    if (pg_wrapper->capture_object_names) {
        for (int i = 0; i < pg_wrapper->num_capture_objects; i++) {
            free(pg_wrapper->capture_object_names[i]);
        }
        free(pg_wrapper->capture_object_names);
    }
    
    free(pg_wrapper);
}

void dlms_result_free(dlms_result_t* result) {
    if (!result) return;
    
    free(result->error_message);
    
    if (result->column_names) {
        for (int i = 0; i < result->num_columns; i++) {
            free(result->column_names[i]);
        }
        free(result->column_names);
    }
    
    if (result->data) {
        int total_cells = result->num_rows * result->num_columns;
        for (int i = 0; i < total_cells; i++) {
            free(result->data[i]);
        }
        free(result->data);
    }
    
    free(result);
}

const char* dlms_result_get_data(dlms_result_t* result, int row, int col) {
    if (!result || !result->data || row < 0 || row >= result->num_rows || 
        col < 0 || col >= result->num_columns) {
        return NULL;
    }
    
    return result->data[row * result->num_columns + col];
}

const char* dlms_result_get_column_name(dlms_result_t* result, int col) {
    if (!result || !result->column_names || col < 0 || col >= result->num_columns) {
        return NULL;
    }
    
    return result->column_names[col];
}

// Helper function to convert access mode to string
static char* access_mode_to_string(DLMS_ACCESS_MODE mode) {
    switch (mode) {
        case DLMS_ACCESS_MODE_NONE:
            return safe_strdup("None");
        case DLMS_ACCESS_MODE_READ:
            return safe_strdup("Read");
        case DLMS_ACCESS_MODE_WRITE:
            return safe_strdup("Write");
        case DLMS_ACCESS_MODE_READ_WRITE:
            return safe_strdup("ReadWrite");
        case DLMS_ACCESS_MODE_AUTHENTICATED_READ:
            return safe_strdup("AuthenticatedRead");
        case DLMS_ACCESS_MODE_AUTHENTICATED_WRITE:
            return safe_strdup("AuthenticatedWrite");
        case DLMS_ACCESS_MODE_AUTHENTICATED_READ_WRITE:
            return safe_strdup("AuthenticatedReadWrite");
        default:
            return safe_strdup("Unknown");
    }
}

// Helper function to convert method access mode to string
static char* method_access_mode_to_string(DLMS_METHOD_ACCESS_MODE mode) {
    switch (mode) {
        case DLMS_METHOD_ACCESS_MODE_NONE:
            return safe_strdup("None");
        case DLMS_METHOD_ACCESS_MODE_ACCESS:
            return safe_strdup("Access");
        case DLMS_METHOD_ACCESS_MODE_AUTHENTICATED_ACCESS:
            return safe_strdup("AuthenticatedAccess");
        case DLMS_METHOD_ACCESS_MODE_AUTHENTICATED_REQUEST:
            return safe_strdup("AuthenticatedRequest");
        case DLMS_METHOD_ACCESS_MODE_ENCRYPTED_REQUEST:
            return safe_strdup("EncryptedRequest");
        case DLMS_METHOD_ACCESS_MODE_DIGITALLY_SIGNED_REQUEST:
            return safe_strdup("DigitallySignedRequest");
        case DLMS_METHOD_ACCESS_MODE_AUTHENTICATED_RESPONSE:
            return safe_strdup("AuthenticatedResponse");
        case DLMS_METHOD_ACCESS_MODE_ENCRYPTED_RESPONSE:
            return safe_strdup("EncryptedResponse");
        case DLMS_METHOD_ACCESS_MODE_DIGITALLY_SIGNED_RESPONSE:
            return safe_strdup("DigitallySignedResponse");
        default:
            return safe_strdup("Unknown");
    }
}

association_view_t* meter_get_association_view(meter_t* meter) {
    if (!meter) {
        association_view_t* result = calloc(1, sizeof(association_view_t));
        if (result) {
            result->error_code = -1;
            result->error_message = safe_strdup("Invalid meter configuration");
        }
        return result;
    }
    
    if (!meter->is_connected || !meter->connection) {
        association_view_t* result = calloc(1, sizeof(association_view_t));
        if (result) {
            result->error_code = -2;
            result->error_message = safe_strdup("Meter not connected. Call meter_connect() first.");
        }
        return result;
    }
    
    connection* con = (connection*)meter->connection;
    int ret;
    
    association_view_t* result = calloc(1, sizeof(association_view_t));
    if (!result) return NULL;
    
    // Get association view using the communication helper
    ret = com_getAssociationView(con, NULL);
    if (ret != 0) {
        result->error_code = ret;
        result->error_message = safe_strdup(hlp_getErrorMessage(ret));
        return result;
    }
    
    // Get the object list from the settings
    objectArray* objectList = &con->settings.objects;
    result->num_objects = objectList->size;
    
    if (result->num_objects > 0) {
        result->objects = calloc(result->num_objects, sizeof(association_object_t));
        if (!result->objects) {
            result->error_code = -1;
            result->error_message = safe_strdup("Memory allocation failed");
            return result;
        }
        
        // Process each object in the association view
        for (int i = 0; i < result->num_objects; i++) {
            gxObject* obj = objectList->data[i];
            association_object_t* assoc_obj = &result->objects[i];
            
            // Set logical name
            char ln[25];
            hlp_getLogicalNameToString(obj->logicalName, ln);
            assoc_obj->logical_name = safe_strdup(ln);
            
            // Set object type and version
            assoc_obj->object_type = obj->objectType;
            assoc_obj->version = obj->version;
            
            // Get attribute count and method count
            assoc_obj->num_attributes = obj_attributeCount(obj);
            assoc_obj->num_methods = obj_methodCount(obj);
            
            // Allocate and set attribute access modes
            if (assoc_obj->num_attributes > 0) {
                assoc_obj->attribute_access_modes = calloc(assoc_obj->num_attributes, sizeof(char*));
                if (assoc_obj->attribute_access_modes) {
                    for (int attr = 0; attr < assoc_obj->num_attributes; attr++) {
                        DLMS_ACCESS_MODE mode = DLMS_ACCESS_MODE_READ_WRITE; // Default
                        if (obj->access && obj->access->attributeAccessModes.size > attr) {
                            mode = (DLMS_ACCESS_MODE)obj->access->attributeAccessModes.data[attr];
                        }
                        assoc_obj->attribute_access_modes[attr] = access_mode_to_string(mode);
                    }
                }
            }
            
            // Allocate and set method access modes
            if (assoc_obj->num_methods > 0) {
                assoc_obj->method_access_modes = calloc(assoc_obj->num_methods, sizeof(char*));
                if (assoc_obj->method_access_modes) {
                    for (int method = 0; method < assoc_obj->num_methods; method++) {
                        DLMS_METHOD_ACCESS_MODE mode = DLMS_METHOD_ACCESS_MODE_ACCESS; // Default
                        if (obj->access && obj->access->methodAccessModes.size > method) {
                            mode = (DLMS_METHOD_ACCESS_MODE)obj->access->methodAccessModes.data[method];
                        }
                        assoc_obj->method_access_modes[method] = method_access_mode_to_string(mode);
                    }
                }
            }
        }
    }
    
    result->error_code = 0;
    result->error_message = safe_strdup("Success");
    
    return result;
}

void association_view_free(association_view_t* view) {
    if (view) {
    if (view->objects) {
        for (int i = 0; i < view->num_objects; i++) {
                if (view->objects[i].logical_name) {
                    free(view->objects[i].logical_name);
                }
                if (view->objects[i].attribute_access_modes) {
                    for (int j = 0; j < view->objects[i].num_attributes; j++) {
                        if (view->objects[i].attribute_access_modes[j]) {
                            free(view->objects[i].attribute_access_modes[j]);
                        }
                    }
                    free(view->objects[i].attribute_access_modes);
                }
                if (view->objects[i].method_access_modes) {
                    for (int j = 0; j < view->objects[i].num_methods; j++) {
                        if (view->objects[i].method_access_modes[j]) {
                            free(view->objects[i].method_access_modes[j]);
                        }
                    }
                    free(view->objects[i].method_access_modes);
                }
            }
            free(view->objects);
        }
        if (view->error_message) {
            free(view->error_message);
        }
        free(view);
    }
} 

// Helper function to parse OBIS code string into byte array
static int parse_obis_code(const char* obis_code, unsigned char* ln) {
    int values[6];
    int count = sscanf(obis_code, "%d.%d.%d.%d.%d.%d", 
                      &values[0], &values[1], &values[2], 
                      &values[3], &values[4], &values[5]);
    
    if (count != 6) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    for (int i = 0; i < 6; i++) {
        if (values[i] < 0 || values[i] > 255) {
            return DLMS_ERROR_CODE_INVALID_PARAMETER;
        }
        ln[i] = (unsigned char)values[i];
    }
    
    return DLMS_ERROR_CODE_OK;
}

// Helper function to send write messages
static int send_write_messages(meter_t* meter, message* messages) {
    if (!meter || !meter->connection || !messages) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    connection* conn = (connection*)meter->connection;
    
    for (uint16_t pos = 0; pos < messages->size; pos++) {
        gxByteBuffer* bb = messages->data[pos];
        
        // Print outgoing packet if debug is enabled
        if (meter->debug_packets) {
            printf("\n[DLMS WRITE DEBUG] Sending message %d/%d\n", pos + 1, messages->size);
            print_packet_hex("OUTGOING", bb->data, bb->size);
        }
        
        if (send(conn->socket, bb->data, bb->size, 0) < 0) {
            if (meter->debug_packets) {
                printf("[DLMS WRITE DEBUG] ERROR: Failed to send packet\n");
            }
            return DLMS_ERROR_CODE_SEND_FAILED;
        }
        
        if (meter->debug_packets) {
            printf("[DLMS WRITE DEBUG] Packet sent successfully, waiting for response...\n");
        }
        
        // Wait for response
        gxByteBuffer reply;
        bb_init(&reply);
        bb_capacity(&reply, 1024);
        
        int bytes = recv(conn->socket, reply.data, reply.capacity, 0);
        if (bytes <= 0) {
            if (meter->debug_packets) {
                printf("[DLMS WRITE DEBUG] ERROR: Failed to receive response (bytes=%d)\n", bytes);
            }
            bb_clear(&reply);
            return DLMS_ERROR_CODE_RECEIVE_FAILED;
        }
        
        reply.size = bytes;
        
        // Print incoming packet if debug is enabled
        if (meter->debug_packets) {
            printf("[DLMS WRITE DEBUG] Received response:\n");
            print_packet_hex("INCOMING", reply.data, reply.size);
        }
        
        gxReplyData replyData;
        reply_init(&replyData);
        
        int ret = cl_getData(&conn->settings, &reply, &replyData);
        
        if (meter->debug_packets) {
            printf("[DLMS WRITE DEBUG] cl_getData returned: %d (%s)\n", 
                   ret, ret == DLMS_ERROR_CODE_OK ? "SUCCESS" : "ERROR");
        }
        
        bb_clear(&reply);
        reply_clear(&replyData);
        
        if (ret != DLMS_ERROR_CODE_OK) {
            return ret;
        }
    }
    
    if (meter->debug_packets) {
        printf("[DLMS WRITE DEBUG] All messages processed successfully\n\n");
    }
    
    return DLMS_ERROR_CODE_OK;
}

// Write int8 value to OBIS code
int meter_write_obis_int8(meter_t* meter, const char* obis_code, int8_t value, int object_type, int attribute_index) {
    if (!meter || !meter->connection || !obis_code) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    unsigned char ln[6];
    int ret = parse_obis_code(obis_code, ln);
    if (ret != DLMS_ERROR_CODE_OK) {
        return ret;
    }
    
    connection* conn = (connection*)meter->connection;
    
    dlmsVARIANT writeValue;
    var_init(&writeValue);
    var_setInt8(&writeValue, value);
    
    message messages;
    mes_init(&messages);
    
    ret = cl_writeLN(&conn->settings, ln, object_type, attribute_index, &writeValue, 0, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        var_clear(&writeValue);
        mes_clear(&messages);
        return ret;
    }
    
    ret = send_write_messages(meter, &messages);
    
    var_clear(&writeValue);
    mes_clear(&messages);
    
    return ret;
}

// Write int16 value to OBIS code
int meter_write_obis_int16(meter_t* meter, const char* obis_code, int16_t value, int object_type, int attribute_index) {
    if (!meter || !meter->connection || !obis_code) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    unsigned char ln[6];
    int ret = parse_obis_code(obis_code, ln);
    if (ret != DLMS_ERROR_CODE_OK) {
        return ret;
    }
    
    connection* conn = (connection*)meter->connection;
    
    dlmsVARIANT writeValue;
    var_init(&writeValue);
    var_setInt16(&writeValue, value);
    
    message messages;
    mes_init(&messages);
    
    ret = cl_writeLN(&conn->settings, ln, object_type, attribute_index, &writeValue, 0, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        var_clear(&writeValue);
        mes_clear(&messages);
        return ret;
    }
    
    ret = send_write_messages(meter, &messages);
    
    var_clear(&writeValue);
    mes_clear(&messages);
    
    return ret;
}

// Write int32 value to OBIS code
int meter_write_obis_int32(meter_t* meter, const char* obis_code, int32_t value, int object_type, int attribute_index) {
    if (!meter || !meter->connection || !obis_code) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    unsigned char ln[6];
    int ret = parse_obis_code(obis_code, ln);
    if (ret != DLMS_ERROR_CODE_OK) {
        return ret;
    }
    
    connection* conn = (connection*)meter->connection;
    
    dlmsVARIANT writeValue;
    var_init(&writeValue);
    var_setInt32(&writeValue, value);
    
    message messages;
    mes_init(&messages);
    
    ret = cl_writeLN(&conn->settings, ln, object_type, attribute_index, &writeValue, 0, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        var_clear(&writeValue);
        mes_clear(&messages);
        return ret;
    }
    
    ret = send_write_messages(meter, &messages);
    
    var_clear(&writeValue);
    mes_clear(&messages);
    
    return ret;
}

// Write uint8 value to OBIS code
int meter_write_obis_uint8(meter_t* meter, const char* obis_code, uint8_t value, int object_type, int attribute_index) {
    if (!meter || !meter->connection || !obis_code) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    unsigned char ln[6];
    int ret = parse_obis_code(obis_code, ln);
    if (ret != DLMS_ERROR_CODE_OK) {
        return ret;
    }
    
    connection* conn = (connection*)meter->connection;
    
    dlmsVARIANT writeValue;
    var_init(&writeValue);
    var_setUInt8(&writeValue, value);
    
    message messages;
    mes_init(&messages);
    
    ret = cl_writeLN(&conn->settings, ln, object_type, attribute_index, &writeValue, 0, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        var_clear(&writeValue);
        mes_clear(&messages);
        return ret;
    }
    
    ret = send_write_messages(meter, &messages);
    
    var_clear(&writeValue);
    mes_clear(&messages);
    
    return ret;
}

// Write uint16 value to OBIS code
int meter_write_obis_uint16(meter_t* meter, const char* obis_code, uint16_t value, int object_type, int attribute_index) {
    if (!meter || !meter->connection || !obis_code) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    unsigned char ln[6];
    int ret = parse_obis_code(obis_code, ln);
    if (ret != DLMS_ERROR_CODE_OK) {
        return ret;
    }
    
    connection* conn = (connection*)meter->connection;
    
    dlmsVARIANT writeValue;
    var_init(&writeValue);
    var_setUInt16(&writeValue, value);
    
    message messages;
    mes_init(&messages);
    
    ret = cl_writeLN(&conn->settings, ln, object_type, attribute_index, &writeValue, 0, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        var_clear(&writeValue);
        mes_clear(&messages);
        return ret;
    }
    
    ret = send_write_messages(meter, &messages);
    
    var_clear(&writeValue);
    mes_clear(&messages);
    
    return ret;
}

// Write uint32 value to OBIS code
int meter_write_obis_uint32(meter_t* meter, const char* obis_code, uint32_t value, int object_type, int attribute_index) {
    if (!meter || !meter->connection || !obis_code) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    unsigned char ln[6];
    int ret = parse_obis_code(obis_code, ln);
    if (ret != DLMS_ERROR_CODE_OK) {
        return ret;
    }
    
    connection* conn = (connection*)meter->connection;
    
    dlmsVARIANT writeValue;
    var_init(&writeValue);
    var_setUInt32(&writeValue, value);
    
    message messages;
    mes_init(&messages);
    
    ret = cl_writeLN(&conn->settings, ln, object_type, attribute_index, &writeValue, 0, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        var_clear(&writeValue);
        mes_clear(&messages);
        return ret;
    }
    
    ret = send_write_messages(meter, &messages);
    
    var_clear(&writeValue);
    mes_clear(&messages);
    
    return ret;
}

// Write float32 value to OBIS code
int meter_write_obis_float32(meter_t* meter, const char* obis_code, float value, int object_type, int attribute_index) {
    if (!meter || !meter->connection || !obis_code) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    unsigned char ln[6];
    int ret = parse_obis_code(obis_code, ln);
    if (ret != DLMS_ERROR_CODE_OK) {
        return ret;
    }
    
    connection* conn = (connection*)meter->connection;
    
    dlmsVARIANT writeValue;
    var_init(&writeValue);
    var_setDouble(&writeValue, (double)value);
    
    message messages;
    mes_init(&messages);
    
    ret = cl_writeLN(&conn->settings, ln, object_type, attribute_index, &writeValue, 0, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        var_clear(&writeValue);
        mes_clear(&messages);
        return ret;
    }
    
    ret = send_write_messages(meter, &messages);
    
    var_clear(&writeValue);
    mes_clear(&messages);
    
    return ret;
}

// Write float64 value to OBIS code
int meter_write_obis_float64(meter_t* meter, const char* obis_code, double value, int object_type, int attribute_index) {
    if (!meter || !meter->connection || !obis_code) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    unsigned char ln[6];
    int ret = parse_obis_code(obis_code, ln);
    if (ret != DLMS_ERROR_CODE_OK) {
        return ret;
    }
    
    connection* conn = (connection*)meter->connection;
    
    dlmsVARIANT writeValue;
    var_init(&writeValue);
    var_setDouble(&writeValue, value);
    
    message messages;
    mes_init(&messages);
    
    ret = cl_writeLN(&conn->settings, ln, object_type, attribute_index, &writeValue, 0, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        var_clear(&writeValue);
        mes_clear(&messages);
        return ret;
    }
    
    ret = send_write_messages(meter, &messages);
    
    var_clear(&writeValue);
    mes_clear(&messages);
    
    return ret;
}

// Write string value to OBIS code
int meter_write_obis_string(meter_t* meter, const char* obis_code, const char* value, int object_type, int attribute_index) {
    if (!meter || !meter->connection || !obis_code || !value) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    unsigned char ln[6];
    int ret = parse_obis_code(obis_code, ln);
    if (ret != DLMS_ERROR_CODE_OK) {
        return ret;
    }
    
    connection* conn = (connection*)meter->connection;
    
    dlmsVARIANT writeValue;
    var_init(&writeValue);
    
    gxByteBuffer bb;
    bb_init(&bb);
    bb_addString(&bb, value);
    var_addOctetString(&writeValue, &bb);
    
    message messages;
    mes_init(&messages);
    
    ret = cl_writeLN(&conn->settings, ln, object_type, attribute_index, &writeValue, 0, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        var_clear(&writeValue);
        bb_clear(&bb);
        mes_clear(&messages);
        return ret;
    }
    
    ret = send_write_messages(meter, &messages);
    
    var_clear(&writeValue);
    bb_clear(&bb);
    mes_clear(&messages);
    
    return ret;
}

// Write datetime value to OBIS code
int meter_write_obis_datetime(meter_t* meter, const char* obis_code, time_t timestamp, int object_type, int attribute_index) {
    if (!meter || !meter->connection || !obis_code) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    unsigned char ln[6];
    int ret = parse_obis_code(obis_code, ln);
    if (ret != DLMS_ERROR_CODE_OK) {
        return ret;
    }
    
    connection* conn = (connection*)meter->connection;
    
    dlmsVARIANT writeValue;
    var_init(&writeValue);
    
    gxtime time;
    time_initUnix(&time, timestamp);
    var_setDateTime(&writeValue, &time);
    
    message messages;
    mes_init(&messages);
    
    ret = cl_writeLN(&conn->settings, ln, object_type, attribute_index, &writeValue, 0, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        var_clear(&writeValue);
        mes_clear(&messages);
        return ret;
    }
    
    ret = send_write_messages(meter, &messages);
    
    var_clear(&writeValue);
    mes_clear(&messages);
    
    return ret;
}

// Write boolean value to OBIS code
int meter_write_obis_boolean(meter_t* meter, const char* obis_code, unsigned char value, int object_type, int attribute_index) {
    if (!meter || !meter->connection || !obis_code) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    unsigned char ln[6];
    int ret = parse_obis_code(obis_code, ln);
    if (ret != DLMS_ERROR_CODE_OK) {
        return ret;
    }
    
    connection* conn = (connection*)meter->connection;
    
    dlmsVARIANT writeValue;
    var_init(&writeValue);
    var_setBoolean(&writeValue, value);
    
    message messages;
    mes_init(&messages);
    
    ret = cl_writeLN(&conn->settings, ln, object_type, attribute_index, &writeValue, 0, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        var_clear(&writeValue);
        mes_clear(&messages);
        return ret;
    }
    
    ret = send_write_messages(meter, &messages);
    
    var_clear(&writeValue);
    mes_clear(&messages);
    
    return ret;
}

// Write octet string value to OBIS code
int meter_write_obis_octet_string(meter_t* meter, const char* obis_code, const unsigned char* data, int length, int object_type, int attribute_index) {
    if (!meter || !meter->connection || !obis_code || !data || length <= 0) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }
    
    unsigned char ln[6];
    int ret = parse_obis_code(obis_code, ln);
    if (ret != DLMS_ERROR_CODE_OK) {
        return ret;
    }
    
    connection* conn = (connection*)meter->connection;
    
    dlmsVARIANT writeValue;
    var_init(&writeValue);
    
    gxByteBuffer bb;
    bb_init(&bb);
    bb_set(&bb, data, length);
    var_addOctetString(&writeValue, &bb);
    
    message messages;
    mes_init(&messages);
    
    ret = cl_writeLN(&conn->settings, ln, object_type, attribute_index, &writeValue, 0, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        var_clear(&writeValue);
        bb_clear(&bb);
        mes_clear(&messages);
        return ret;
    }
    
    ret = send_write_messages(meter, &messages);
    
    var_clear(&writeValue);
    bb_clear(&bb);
    mes_clear(&messages);
    
    return ret;
} 

/*******************************************************************************
 * Method Invocation Functions
 ******************************************************************************/

static int send_method_request_and_get_reply(meter_t* meter, message* messages) {
    if (!meter || !meter->connection || !messages) {
        return DLMS_ERROR_CODE_INVALID_PARAMETER;
    }

    connection* conn = (connection*)meter->connection;
    int ret = 0;

    for (uint16_t pos = 0; pos < messages->size; pos++) {
        gxByteBuffer* bb = messages->data[pos];

        if (meter->debug_packets) {
            printf("\n[DLMS METHOD CALL DEBUG] Sending message %d/%d\n", pos + 1, messages->size);
            print_packet_hex("OUTGOING (METHOD)", bb->data, bb->size);
        }

        if (send(conn->socket, bb->data, bb->size, 0) < 0) {
            if (meter->debug_packets) {
                printf("[DLMS METHOD CALL DEBUG] ERROR: Failed to send packet\n");
            }
            return DLMS_ERROR_CODE_SEND_FAILED;
        }

        if (meter->debug_packets) {
            printf("[DLMS METHOD CALL DEBUG] Packet sent successfully, waiting for reply...\n");
        }

        gxByteBuffer reply;
        bb_init(&reply);
        bb_capacity(&reply, 1024);

        int bytes = recv(conn->socket, reply.data, reply.capacity, 0);
        if (bytes <= 0) {
            if (meter->debug_packets) {
                printf("[DLMS METHOD CALL DEBUG] ERROR: Failed to receive reply (bytes=%d)\n", bytes);
            }
            bb_clear(&reply);
            return DLMS_ERROR_CODE_RECEIVE_FAILED;
        }
        reply.size = bytes;

        if (meter->debug_packets) {
            printf("[DLMS METHOD CALL DEBUG] Received reply:\n");
            print_packet_hex("INCOMING (METHOD)", reply.data, reply.size);
        }

        gxReplyData replyData;
        reply_init(&replyData);
        ret = cl_getData(&conn->settings, &reply, &replyData);

        bb_clear(&reply);
        if (ret != DLMS_ERROR_CODE_OK) {
            reply_clear(&replyData);
            return ret;
        }
        
        if(replyData.command == DLMS_COMMAND_METHOD_RESPONSE)
        {
            if (replyData.data.size > 0)
            {
                ret = replyData.data.data[0];
            }
        }

        reply_clear(&replyData);
        if (ret != DLMS_ERROR_CODE_OK) {
            return ret;
        }
    }

    if (meter->debug_packets) {
        printf("[DLMS METHOD CALL DEBUG] All messages processed successfully\n\n");
    }

    return DLMS_ERROR_CODE_OK;
}

int meter_call_method_no_params(meter_t* meter, const char* obis_code, int object_type, int method_index) {
    return meter_call_method_with_data(meter, obis_code, object_type, method_index, NULL, 0);
}

int meter_call_method_with_data(meter_t* meter, const char* obis_code, int object_type, int method_index, const unsigned char* data, int data_length) {
    if (!meter || !meter->is_connected) {
        return DLMS_ERROR_CODE_NOT_INITIALIZED;
    }

    unsigned char ln[6];
    if (parse_obis_code(obis_code, ln) != 0) {
        return DLMS_ERROR_CODE_INVALID_LOGICAL_NAME;
    }

    connection* conn = (connection*)meter->connection;
    message messages;
    mes_init(&messages);

    int ret = cl_methodLN2(&conn->settings, ln, object_type, method_index, (unsigned char*)data, data_length, &messages);
    if (ret != DLMS_ERROR_CODE_OK) {
        mes_clear(&messages);
        return ret;
    }

    ret = send_method_request_and_get_reply(meter, &messages);
    mes_clear(&messages);
    return ret;
} 

int meter_call_set_time(meter_t* meter, time_t timestamp)
{
    if (!meter || !meter->is_connected)
    {
        return DLMS_ERROR_CODE_NOT_INITIALIZED;
    }
    const char *obis_code_str = "0.0.1.0.0.255";
    unsigned char ln[6];
    if (parse_obis_code(obis_code_str, ln) != 0)
    {
        return DLMS_ERROR_CODE_INVALID_LOGICAL_NAME;
    }
    DLMS_OBJECT_TYPE object_type = DLMS_OBJECT_TYPE_CLOCK;
    int method_index = 7;
    gxtime dt;
    time_initUnix(&dt, (uint32_t)timestamp);
    gxByteBuffer time_bytes;
    bb_init(&time_bytes);
    int ret = cosem_setDateTimeAsOctetString(&time_bytes, &dt);
    if (ret != DLMS_ERROR_CODE_OK)
    {
        bb_clear(&time_bytes);
        return ret;
    }
    dlmsVARIANT data;
    var_init(&data);
    var_addOctetString(&data, &time_bytes);
    connection *conn = (connection *)meter->connection;
    message messages;
    mes_init(&messages);
    ret = cl_methodLN(&conn->settings, ln, object_type, method_index, &data, &messages);
    bb_clear(&time_bytes);
    var_clear(&data);
    if (ret != DLMS_ERROR_CODE_OK)
    {
        mes_clear(&messages);
        return ret;
    }
    ret = send_method_request_and_get_reply(meter, &messages);
    mes_clear(&messages);
    return ret;
} 