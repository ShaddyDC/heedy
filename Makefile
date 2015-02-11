.PHONY: clean all dependencies test

#gets the list of files that we're to compile
SRC=$(wildcard src/core/*.go)

#Get the list of executables from the file list
TMPO=$(patsubst src/core/%.go,bin/%,$(SRC))
OBJ=$(TMPO:.go=)



all: clean dependencies $(OBJ) bin/dep/gnatsd test

build: $(OBJ)

bin:
	mkdir bin
	mkdir bin/dep
	mkdir bin/test_coverage
	cp -r config bin/config

#Rule to go from source go file to binary
bin/%: src/core/%.go bin
	go build -o $@ $<

clean:
	rm -rf bin
	go clean

############################################################################################################
#Dependencies of the project
############################################################################################################

dependencies:
	go get github.com/apcera/nats
	go get github.com/apcera/gnatsd
	go get github.com/garyburd/redigo/redis
	go get github.com/mattn/go-sqlite3
	go get github.com/nu7hatch/gouuid
	go get github.com/gorilla/mux
	go get github.com/gorilla/context
	go get gopkg.in/mgo.v2
	go get github.com/gorilla/sessions

#gnatsd is the messenger server - deps must be installed, but we don't want deps to be called
#each time we check for gnatsd executable or each time tests are run
bin/dep/gnatsd: bin
	go build -o bin/dep/gnatsd github.com/apcera/gnatsd

############################################################################################################
#Running Tests
############################################################################################################

test: bin/dep/gnatsd
	./runtests.sh
