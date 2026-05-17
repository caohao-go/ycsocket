# Change Log

## v0.0.11 (2026-03-04)

### Features

- websocket: support client handshake headers (!264)

## v0.0.10 (2025-03-05)

### Features

- Add combined writes optimization to merge header and payload into a single syscall (!249)

## v0.0.9 (2024-09-25)

### Features

- support {read|write} idletimeout (!240)

## v0.0.8 (2024-09-03)

### Bug Fixes

- Upgrade tnet to v0.0.19 to avoid oom when desc is not freed promptly (!224)

## v0.0.7 (2024-06-26)

### Features

- Support ReadAnyMessage (!227)
- Support asynchronous ping/pong handler (!225)

## v0.0.6 (2024-02-02)

### Documentation

- Update description of the websocket.Conn API (!220)

## v0.0.5 (2023-09-19)

### Features

- Upgrade tnet to v0.0.16 (!209)

## v0.0.4 (2023-08-31)

### Features

- Upgrade tnet to v0.0.15 (!207)

## v0.0.3 (2023-04-18)

### Features

- Provides lock for concurrent write (!175)
- Return message type for errors (!176)
- Provide with handshake context (!174)
- Add local address into context (!173)
- Add remote address into context before upgrade (!172)
- Support client set on request handle (!155)

## v0.0.2 (2022-12-08)

### Features

- Add before handshake hook, register ws's onClose into tnet to ensure that it is called on a passive close (!148)

## v0.0.1 (2022-05-26)

### Features

- Read/Write for a full message
- NextMessageReader/NextMessageWriter for user customized read/write
- Writev for multiple byte slices at server side
- SetMetadata/GetMetadata to store/retrieve user's private data
- Customized control frame handler for Ping/Pong/Close
- Set the message type of the connection to use Read/Write API directly
- Provide examples and readme
- Provide both client and server programming APIs 
