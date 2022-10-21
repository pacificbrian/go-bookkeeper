# SPDX-FileCopyrightText: 2022 Brian Welty
#
# SPDX-License-Identifier: MPL-2.0
#
# simple Makefile to assist with go/yarn/webpack
#

build: client server

first-time:
	brew install golang yarn

deps:
	go get .
	yarn install

upgrade-deps:
	go get -u
	yarn upgrade

js:
	yarn run webpack --config javascript/webpack.dev.js

js-prod:
	yarn run webpack --config javascript/webpack.prod.js

client: js-prod

server:
	go build

install: deps build
	go install
