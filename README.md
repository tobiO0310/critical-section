# Ricart-Agrawala Critical Section

Please see the section below on how to run the servers.

The servers use ports 3000-3009, so make sure nothing use those as well.

You must have downloaded the repository and have one terminal open, with the working directory being the git root.

## Servers
To run the servers, make sure to have golang, gRPC, protobuf, and Make (GNU) installed.
To run the servers, run the following command in your terminal:
```bash
$ make run
```
To view the behavior of a given server, check the log file `output_server_\<id\>.log` where '\<id\>' is the server's id.

To stop running the servers, exit using `ctrl-c`
