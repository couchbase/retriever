retriever
=========

A dog or breed used for retrieving game such as log files and process stats

1. logger - Error logger. Can log data on the local machine or to a remote machine. 

2. stats  stats collection and reporting. 


HOWTO
------

Start retriver process
./retriver

Start the example_server and example_client process

List of supported commands

1. Set the log level dynamically
For module ExampleServer
curl -v -i -X POST -d '{"Cmd":"level", "Message":"debug"}' http://localhost:8080/logger/ExampleServer
For all modules
curl -v -i -X POST -d '{"Cmd":"level", "Message":"warn"}' http://localhost:8080/logger/All

2. get log 
curl -v -i -X POST -d '{"Cmd":"log"}' http://localhost:8080/logger/ExampleServer

curl -v -i -X POST -d '{"Cmd":"log"}' http://localhost:8080/logger/all

3. Enable/disable transaction logging

curl -v -i -X POST -d '{"transEnable"}' http://localhost:8080/logger/all 

curl -v -i -X POST -d '{"transDisable"}' http://localhost:8080/logger/all

4. get transaction log 

curl -v -i -X POST -d '{"Cmd":"transactionLog", "Message":"1004320"}' http://localhost:8080/logger/test


5. Get all transaction logs 
curl -v -i -X POST -d '{"Cmd":"transactionLog"}' http://localhost:8080/logger/all

6. Log Rotate for module

curl -v -i -X POST -d '{"Cmd":"rotate"}' http://localhost:8080/logger/ExampleServer

7. Log Rotate for all modules
curl -v -i -X POST -d '{"Cmd":"rotate"}' http://localhost:8080/logger/all

8. Stats
curl -v -i http://localhost:8080/stats/ExampleServer

curl -v -i http://localhost:8080/stats/all
