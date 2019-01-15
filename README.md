# Wormhole Engine

The project used to sync wormhole node data to database.
## What is Wormhole

Wormhole is a fast, portable Omni Layer implementation that is based off the Bitcoin ABC codebase (currently 0.18.2). This implementation requires no external dependencies extraneous to Bitcoin ABC, and is native to the Bitcoin Cash network just like other Bitcoin Cash nodes. It currently supports a wallet mode and is seamlessly available on three platforms: Windows, Linux and Mac OS. Wormhole Cash Layer extensions are exposed via the JSON-RPC interface. Development has been consolidated on the Wormhole product, and it is the reference client for the Wormhole Cash Layer.

## Quick Start

#### Prerequesites
| Package | Version | 
| :------| ------|
| Mysql | 5.7+ |
| Golang | 1.10+ |
| Redis | 4.0+ |
|[Wormhole](https://github.com/copernet/wormhole)|0.2.2|

#### Database Init

1、Once the mysql installation is complete.you need create your database first.

	#connect your database
	mysql -u{your-user} -p{your-password} -h{your-host} -P{your-port}
	
	#create database
	mysql> create database wormhole;
2、Init tables

	go get -d github.com/copernet/whccommon/model/operation/setup
	cd ${gopath}/src/github.com/copernet/whccommon/model/operation/setup/
	go build
	./setup --host={your-host} --port={your-port} --user={your-user} --passwd={your-password} --db=wormhole --walletdb=-

#### Config Init

	go get -d github.com/wormholeSV/whcengine
	cd ${gopath}/src/github.com/wormholeSV/whcengine
	cp conf.yml.sample conf.yml

	#you need modify db、redis、rpc、log to your local config
#### How To Run

	cd ${gopath}/src/github.com/wormholeSV/whcengine
	mkdir logs
	touch logs/system.log
	go build

	#start
	tools/run start whcengine
	#stop 
	tools/run stop whcengine

