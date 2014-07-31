retriever
=========

A dog or breed used for retrieving game such as log files and process stats. Retriever is a 
highly flexible and configurable logging & stats package that allows you to control a bunch 
of logging options via a REST interface

1. logger - Error logger. Can log data on the local machine or to a remote machine. 

2. stats  stats collection and reporting. 

Windows build supported
-----------------------

After checkout (uses github.com/natefinch/npipes) run
go get ./...

HOWTO
------

Start retriver process
./retriver

Start the example_server and example_client process

List of supported commands

------
Set the log level dynamically

For module ExampleServer

curl -v -i -X POST -d '{"Cmd":"level", "Message":"debug"}' http://localhost:8080/logger/ExampleServer

For all modules

curl -v -i -X POST -d '{"Cmd":"level", "Message":"warn"}' http://localhost:8080/logger/All

------
Retrieve Logs

curl -v -i -X POST -d '{"Cmd":"log"}' http://localhost:8080/logger/ExampleServer

curl -v -i -X POST -d '{"Cmd":"log"}' http://localhost:8080/logger/all

------
Enable/disable trace logging

curl -v -i -X POST -d '{"traceEnable"}' http://localhost:8080/logger/all 

curl -v -i -X POST -d '{"traceDisable"}' http://localhost:8080/logger/all

------
get trace log for ExampleServer 

curl -v -i -X POST -d '{"Cmd":"transactionLog", "Message":"1004320"}' http://localhost:8080/logger/ExampleServer

------
Get all  trace logs

curl -v -i -X POST -d '{"Cmd":"transactionLog"}' http://localhost:8080/logger/all

------
Log Rotate for ExampleServer

curl -v -i -X POST -d '{"Cmd":"rotate"}' http://localhost:8080/logger/ExampleServer

------
Log Rotate for all processes

curl -v -i -X POST -d '{"Cmd":"rotate"}' http://localhost:8080/logger/all

------
Change the default logging path for a module (Not supported for "All" )

curl -v -i -X POST -d '{"Cmd":"path", "Message":"/dev/shm"}' http://localhost:8080/logger/ExampleServer

------
Configure a remote server for sending alerts. Only error messages are sent

curl -v -i -X POST -d '{"Cmd":"alarmSet", "Message": "http://localhost:9111/alarm/"}' http://localhost:8080/logger/all

For a single module

curl -v -i -X POST -d '{"Cmd":"alarmSet", "Message": "http://localhost:9111/alarm/"}' http://localhost:8080/logger/ExampleServer

Disable Alerts
curl -v -i -X POST -d '{"Cmd":"alarmClear"}' http://localhost:8080/logger/all

------
Stats

curl -v -i http://localhost:8080/stats/ExampleServer

curl -v -i http://localhost:8080/stats/all
