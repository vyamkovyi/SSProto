# ServerSync protocol (Version 2)

## Communication

Is done over with TLS over TCP on port 48879.
All binary data is serialized in little-endian format.

## Dynamic-length data

If a client or server needs to send binary data without known predefined length
(like, UUIDs are always 32 bytes) it sends length (64-bit unsigned integer)
before sending data itself.

```
+-----------------+-- .... --+
| length (uint64) |   data   |
+-----------------+-- .... --+
```

### Filesystem paths

All filesystem paths sent through network MUST use forward slash (`/`) as a
path component separator. Both sides MUST implement corresponding translation 
to/from OS-specific format.

## Typical session

#### Stage 0: Preparation

1. Both sides send protocol version (8-bit unsigned integer) to another side.
   If they don't match, the client MUST close the connection.

2. Client sends it's unique 32-byte identifier.

3. The server replies either with 1 or 0 (8-bit unsigned integer).
   If value is 0 - update request is "rejected" and 
   server closes connection. The client MUST consider the update
   to be successful in this case.

3. The client sends information about its hardware (dynamic-length)
   This protocol doesn't define any requirements for its format, but
   the current implementation uses JSON-encoded blob with OS id and 
   memory usage statistic.

### Stage 1: Client file list sending

1. The client sends hash-list entry (see below) which describes a separate
   file in a client-side file tree. The client MAY ignore (not send entries for)
   some files during this stage.

Hash-list entry format:
```
+------------------+---------------+--------- .... ---------+
|    file hash     | file path len |          file          |
|    (32 byte)     |    (uint64)   |          path          |
+------------------+---------------+--------- .... ---------+
```
**Hash is 256-bit BLAKE2b.**

2. Server replies with 1 or 0 (8-bit unsigned integer).
   If the value is 0 - file to which entry refers should be deleted.
   The client MAY ignore this and not delete the file even if the server reply is 1.

3. Steps and 1 and 2 are repeated until the client sends 32 zero bytes during step 1.
   In this case, client doesn't send the rest of hash-list entry and server
   proceeds to stage 2.

### Stage 2: Files downloading

The server sends files to the client that should be replaced (or missing).

1. The server sends files in form of special update packets (see format below) and
   then closes the connection.

File blob format:
```
+---------------+----- .... -----+--------------+------- .... -------+
| file path len |      file      |   file size  |        file        |
|    (uint64)   |      path      |   (uint64)   |      contents      |
+---------------+----- .... -----+--------------+------- .... -------+
```
