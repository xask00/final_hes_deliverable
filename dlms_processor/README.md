# About me
DLMS processor is a stateless component whose job in life is to connect to a energy meter and get data from it.
It store anything in a database, it does not remember what was the previous requests that were made.
Every request is treated as a fresh request.

# Main Idea
 - stateless
 - connect to energy meters and do DLMS operations
 - DLMS operations it supports
    - READ OBIS codes
    - WRITE attributes to a OBIS code
    - execute functions on a OBIS code
 - perform BULK operations
