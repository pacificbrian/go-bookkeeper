# simple Makefile to assist with go/yarn/webpack

server:
	go build

deps:
	go get .
	yarn install

upgrade:
	go get -u
	yarn upgrade

js:
	yarn run webpack --config webpack.dev.js

client: js

all: client server
