# Ricart-Agrawala Critical Section

Please see the section below on how to run the servers.

> [!IMPORTANT]
> The service expects 10 servers to be up,
> using the ports 3000-3009.
>
> Please make sure all of these ports are available to be used,
> and that the current user can listen on them.

You must have downloaded the repository and have one terminal open, with the working directory being the git root.

## Servers

To run the servers, make sure to have golang, and optionally Make (GNU) installed.

> [!NOTE]
> You may install protobuf and gRPC, if you wish to recompile the protobuf yourself.

To run the servers, run the following command in your terminal:

```bash
$ make run
```

> [!NOTE]
> If you wish to use Make (GNU), please make sure your shell can handle the Make command.
> The shell you run in, must support Bash-style subshells and
> that ampersand (&) means the command runs in the background.
>
> Modern shells like Bash, zsh, sh, etc. support these,
> but notable Fish and Powershell does not support one or both of those requirements.

However, if you cannot use Make you can run the following command in 10 different shells, subshells included.
Please make sure to set the environment variable `CS_ID` to a different value in the range [0; 9] for each (sub)shell.

```bash
$ go run server/*.go
```

A full command, using bash subshells and ampersand for background running, is like so:

```bash
$ (CS_ID="0" go run server/*.go) & (CS_ID="1" go run server/*.go) &\
  (CS_ID="2" go run server/*.go) & (CS_ID="3" go run server/*.go) &\
  (CS_ID="4" go run server/*.go) & (CS_ID="5" go run server/*.go) &\
  (CS_ID="6" go run server/*.go) & (CS_ID="7" go run server/*.go) &\
  (CS_ID="8" go run server/*.go) & (CS_ID="9" go run server/*.go)
```

### View output

To view the behavior of a given server, check the log file `output_server_\<id\>.log` where '\<id\>' is the server's id.

### Finished

Once all nodes has connected to the critical section once, they will continue to listen,
and will only listen. Therefore, you can freely send a terminal signal (e.g. via Ctrl-C) to the processes.

If you are using the make command or the full command,
you can simply do Ctrl-C and it will propagate to all subshells.
