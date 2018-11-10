# ServerSync protocol (Version 2)

## Communication

Is done over with TCP with TLS on port 48879.
All binary data is serialized in little-endian format.

## Dynamic-length data

If client or server needs to send binary data without known predefined length
(like, UUIDs are always 32 byte) it sends length (64-bit unsigned integer)
before sending data itself.

```
+-----------------+-- .... --+
| length (uint64) |   data   |
+-----------------+-- .... --+
```

## Typical session

#### Stage 0: Preparation

1. Both sides send protocol version (8-bit unsigned integer) to other side.
   If they don't match, client MUST close connection.

2. Client sends it's unique 32-byte identifier.

3. Server replies with 1 or 0 (8-bit unsigned integer).
   If value is 0 - update request is "rejected" and 
   server closes connection. Client SHLOUD consider update
   to be "successful" in this case.

3. Client sends informatio about it's hardware (dynamic-length)
   Protocol doesn't defines any requirements for it's format, but
   current implementation uses JSON-encoded blob with OS id and 
   memory usage statistic.

### Stage 1: Client file list sending

1. Client sends hash-list entry (see below) which describes separate file in client-side file tree.
   Client MAY ignore (not send entries for) some files during this stage.

Hash-list entry format:
```
+------------------+---------------+--------- .... ---------+
|    file hash     | file path len |          file          |
|    (32 byte)     |    (uint64)   |          path          |
+------------------+---------------+--------- .... ---------+
```
**Hash is 256-bit BLAKE2b.**

2. Server replies with 1 or 0 (8-bit unsigned integer).
   If value is 0 - file to which entry refers should be deleted.
   Client MAY ignore this and not delete file even if server reply is 1.

3. Steps and 1 and 2 are repeated until client sends 32 zero bytes during step 1.
   In this case client doesn't sends the rest of hash-list entry and server
   proceeds to stage 2.

### Stage 2: Files downloading

Server sends files to client that should be replaced (or missing).

1. Server sends file blobs (see format below) and then closes connection.

File blob format:
```
+---------------+----- .... -----+--------------+------- .... -------+
| file path len |      file      |   file size  |        file        |
|    (uint64)   |      path      |   (uint64)   |      contents      |
+---------------+----- .... -----+--------------+------- .... -------+
```
