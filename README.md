# OnlineVoting
This project can perform elections and votes on different workstations, as long as they can be connected through the network.
This project should be executed in the linux environment.

We have a docker compose file to run the following containers:
1. mongo1
2. mongo2
3. mongo3
4. server1
5. server2
6. client1
```
docker compose up -d
```

Initialize the mongo db replica
```
docker exec mongo1 /scripts/init_db.sh
```

Both server and client have golang environment installed.

## Server side
enter the server container
```
docker exec -it server{1 or 2} bash
```
init the envirionment
```
./init_env.sh
```
Then go to the server folder, execute the following command:
```
go run main.go
```
Then the server will execute. 

## Client side
enter the server container
```
docker exec -it client1 bash
```
init the envirionment
```
./init_env.sh
```
If you want to execute client program, go to the client folder, and execute the command:
```
go run main.go
```
At this point, the client can send a request to the server.
You can perform corresponding tasks through the command line interface provided by the program.
