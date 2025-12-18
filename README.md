# netfs
netfs (Network File System) allows you to manage the file system of computers on the same network.

## Architecture

### Simple architecture of netfs.
```
                ----------------------------------
                |          LOCAL NETWORK         |
                |--------------------------------|
| CLIENT |  ->  | COMPUTER | COMPUTER | COMPUTER |
                |----------|----------|----------|
                |   HOST   |   HOST   |  HOST    |
                ----------------------------------

LOCAL NETWORK - It's a network of computers, for example, your wifi network.
COMPUTER      - It's any PC with the Windows, Linux or MacOS operating system.
HOST          - It's netfs server.
CLIENT        - It's any interface for interacting with netfs.
```

### Simple architecture of netfs host.
```
---------------------------
|         HOST            |
|-------------------------|
|       TRANSPORT         |
|-------------------------| 
|   TASKS   |    VOLUME   |
|-------------------------|
|        DATABASE         |
|-------------------------|
|      FILE SYSTEM        |
---------------------------

HOST        - It's netfs server.
TRANSPORT   - It's an abstraction for interacting with the server can be implemented via HTTP, GRPC, or other protocols.
TASKS       - It's a pool of asynchronous tasks for performing long-term operations, such as copying files.
VOLUME      - It's abstraction for interacting with file system of the computer.
DATABASE    - It's storage for long-term data storing.
FILE SYSTEM - It's file system of the computer.
```