# SPDX-FileCopyrightText: 2022 Brian Welty
#
# SPDX-License-Identifier: MPL-2.0
#
# simple Makefile to assist with go/yarn/webpack
#

build: client server

first-time:
	brew install golang yarn

first-time-ubuntu:
	# first need to add yarn apt repository
	curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
	sudo sh -c 'echo "deb https://dl.yarnpkg.com/debian/ stable main" >> /etc/apt/sources.list.d/yarn.list'
	sudo apt install golang-go yarn

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
