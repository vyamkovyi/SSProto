# SSProto

SSProto is a fairly simple but flexible and secure TCP protocol for deploying
game client updates. Initially, it was developed as a part of Hexamine project
(a Minecraft server) but turns out to be useful outside of Minecraft
environment. You may read about it more on
[Hexawolf's blog](https://hexawolf.me/blog/ssproto/)

## What?

This applications connects to a server (ss-server) that implements SSProto,
"indexes" (hashes) files in current directory and sends list of files and their
hashes to the server. After doing so, it listens to "update packets" from the
server. They contain files, their paths and signatures, that client need to
download and save.

Basically, this application tries to keep working directory in synchronization
with server. For example, if some essential game files were modified on the
server side, this client will download them from the server and game will be
compatible with it's game server again.

## Why?

SSProto is a really simple, hand-made protocol that was made because of all the
problems that appeared with deploying [Hexamine](https://hexawolf.me/hexamine)
updates. Mainly, because people found installing updates (that were deployed
almost every day those days) too painful or even faced problems due to improper
update as a result of customized client setup. While ss-client is tuned
specifically for Hexamine client, it is obvious that SSProto (as well as this
program) may be tuned for more applications where it is essential to keep only
specific server files in sync with clients.

All communications are secured using hardcoded TLS key.

## License

Copyright © 2018 Hexawolf

This software and most of it's code (except utils.go file) is available under
MIT license. The utils.go is not licensed at all: you are hereby granted to do
anything you want with the code there.

