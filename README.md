# ss-server

ss-server is a default implementation of SSProto update protocol used by Hexamine server.

## How does it works?

After setting up TCP connection, client sends it's generated identifier to the server. Server records it but does not
saves to the logs, he sends ed448-signed identifier back to confirm his identity. After that, server and client agree on
SSProto version: if it differs, client must close connection and suggest user to update the updater application.

After doing all the setup work, server expects list of file paths and their hashes. After receiving, it generates a
difference between server and client file indexes. Resulting difference is being sent file-by-file to the client, all
files are signed with ed448-decaf key hard-coded to the client.

## Copyright

Copyright (C) 2018  Hexawolf.
This work uses MIT license.