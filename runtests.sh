#!/bin/bash
echo "Waiting for servers to start..."

DBDIR="database_test"

if [ "$EUID" -eq 0 ]
  then echo "Do not run this script as root!"
  exit 1
fi


if [ -d "$DBDIR" ]; then
    rm -rf $DBDIR
fi

echo "Setting up environment..."

./bin/connectordb create $DBDIR
./bin/connectordb start $DBDIR

echo "Looking for postgres proc..."
ps -aux | grep postgres


echo "Looking for redis proc..."
ps -aux | grep redis-server


echo "Looking for gnatsd proc..."
ps -aux | grep gnatsd

echo "Running tests..."
go test -cover streamdb/...
test_status=$?

./bin/connectordb stop $DBDIR
#rm -rf $DBDIR


if [ $test_status -eq 0 ]; then
	#Now test the python stuff, while rebuilding the db to make sure that
	#the go tests didn't invalidate the db
	./bin/connectordb create --user test:test $DBDIR
    ./bin/connectordb start $DBDIR
    ./bin/restserver --sql=postgres://127.0.0.1:52592/connectordb?sslmode=disable &
    rest_server=$!
    echo "Starting connectordb api tests..."
    nosetests src/clients/python/connectordb_test.py
    test_status=$?
    kill $rest_server
	./bin/connectordb stop $DBDIR
fi


#rm -rf $DBDIR
exit $test_status
