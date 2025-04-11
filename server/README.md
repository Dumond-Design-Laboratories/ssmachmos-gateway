# SSMachMos

Sensor gateway server written in golang.

Entry point is at cmd/ssmachmos/main.go, which in turn:
1. Init server at server/server.go#Init
2. Receives Unix socket messages at api/api.go#Start

## IPC protocol

Server is mostly reply-only. It only returns a response to GUI if asked for one. User messages are uppercase hyphenated verbs with space-separated arguments. Each message must be terminated with a zero byte. 

Server replies are three parts, separated by a colon (`:`) each. First part is message tag, such as `OK`, `ERR`, `MSG`. Second part is the verb the server is replying to, such as `LIST`, `PAIR-LIST`, or `PAIR-ACCEPT`. Third part, if present, is a JSON string. Terminated with a zero byte too.
