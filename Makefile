# Makefile to help uploading prune-cluster scripts (and running it)
# @author: jakob@gedefar.dk

HOST := <HOST-name>

.PHONY: caymap-linux.test
caymap-linux.test: .*
	env GOOS=linux GOARCH=amd64 go test -c -o caymap-linux.test

upload: caymap-linux.test
	scp caymap-linux.test $(HOST):/home/jakob/caymap.test

.PHONY: caymap.test
caymap.test: .*
	go test -c -o caymap.test

local: caymap.test
	# Empty


bench: caymap.test
	./caymap.test -test.run=^\$$ -test.cpu 1 -test.bench '.*'
	#'.*get/same_keys.*'

# Command:  ./caymap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*get/same_keys.*'
ssh:
	@ssh $(HOST)
