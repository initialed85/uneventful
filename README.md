# uneventful

Not sure yet.

## What is it?

Just some tinkering as I try to learn about event based systems- I'm trying to head towards CQRS but I also wanna develop a generalised
pub/sub event framework in the process.

## Concepts

- Domain
    - A slice of functionality ideally centered around something real-world (e.g. `wallet`)
- Entity
    - An instance of something within a domain (e.g. `wallet.28sshuU4BSZ2RCJyTHt2CS5yVeQ`)
- Server
    - The user-facing HTTP entrypoint into a domain (e.g. `http://wallet_server/wallet/28sshuU4BSZ2RCJyTHt2CS5yVeQ/balance`)
- Writer
    - The write path for an entity in a domain (e.g. `nats://message_broker:4222/event.wallet.28sshuU4BSZ2RCJyTHt2CS5yVeQ.credit`)
- Reader
    - The read path for an entity in a domain (e.g. `redis://cache:6379/event.wallet.28sshuU4BSZ2RCJyTHt2CS5yVeQ.balance`)

## TODO

- Get domain writers to push their starting state to Redis
- Get domain writers to re-achieve their state on bootup by replaying their events
- Add /healthz endpoint for all services

## Prerequisites

- Go 1.17+
- Docker
- Docker Compose
- Redis CLI
- cURL
- jq

And optionally for the utilities / integration tests:

- Python3.9+
- Virtualenv

## Usage

### Pull and build

```shell
./pull.sh && ./build.sh && ./run.sh
```

### Run

Foreground

```shell
./run.sh
```

Background

```shell
./run_in_background.sh
```

### Interact

Assuming you've got everything up and running, open a bunch of shells as follows...

#### Shell 1 - Read state for Wallet domain (from Redis)

```shell
while true; do clear; redis-cli GET wallet.28skwt5B8zTrs6AqBWrSgCHLcRL | jq; sleep 1; done
```

#### Shell 2 - Write event log for Wallet domain (from SQLite)

```shell
while true; do clear; ./docker-compose.sh exec wallet_writer_service sqlite3 /var/lib/sqlite/data/datastore.db -line 'SELECT * FROM event;'; sleep 1; done
```

#### Shell 3 - Write state log for Wallet domain (from SQLite)

```shell
while true; do clear; ./docker-compose.sh exec wallet_writer_service sqlite3 /var/lib/sqlite/data/datastore.db -line 'SELECT * FROM state;'; sleep 1; done
```

#### Shell 4 - Balance (user-facing read state) for Wallet domain

```shell
while true; do clear; curl -s http://localhost/wallet/28skwt5B8zTrs6AqBWrSgCHLcRL/balance | jq; sleep 1; done
```

#### Shell 5 - Transactions (user-facing read state) for Wallet domain

```shell
while true; do clear; curl -s http://localhost/wallet/28skwt5B8zTrs6AqBWrSgCHLcRL/transactions | jq; sleep 1; done
```

#### Shell 6 - Let's cause some changes

A debit attempt should fail due to zero balance

```shell
curl -s -X POST -d '{"amount": 5}' http://localhost/wallet/28skwt5B8zTrs6AqBWrSgCHLcRL/debit | jq
```

A credit attempt should increase the balance

```shell
curl -s -X POST -d '{"amount": 5}' http://localhost/wallet/28skwt5B8zTrs6AqBWrSgCHLcRL/credit | jq
```

And now the same debit attempt should succeed

```shell
curl -s -X POST -d '{"amount": 5}' http://localhost/wallet/28skwt5B8zTrs6AqBWrSgCHLcRL/debit | jq
```

Observations in the other shell windows:

- Redis is updated with state for readers to use
- Event log contains all attempted transactions
- State log contains the state after each transaction
- Balance updates appropriately
- Transactions update appropriately

If you really wanna pummel the system, set yourself up with a Virtualenv, install `requirements.txt` and run the following:

```shell
python -m utils.curl -s --loop --workers 64 --period 0 http://localhost/wallet/28skwt5B8zTrs6AqBWrSgCHLcRL/balance
```

This will spin up 64 threads all requesting the balance as fast as they can- I get maybe 250 to 350 requests per second on my Macbook Pro
and as far as I can tell, the system is the limiting factor- attempts to add more workers or more entire instances of that command
see requests per second reduced.

Not sure where the bottleneck is- no single container is working particularly hard, maybe it's just Docker for Mac things.

### How does it work?

#### Service breakdown

- `message_broker` = NATS for pub/sub glue
- `cache` = Redis for caching read state
- `history_writer_datastore` = SQLite for storing global event logs
- `history_writer_service` = Go code to record global write events
- `wallet_writer_datastore` = SQLite for storing event logs and state logs
- `wallet_writer_service` = Go code to handle write events
- `wallet_server_service` = Go code to expose read state

#### Overview

- `wallet_server_service` is really just a convenience abstraction to expose the reader and writer via HTTP
- All reads ultimately happen against Redis
- All writes are handled by `wallet_writer_service`, which ensures to:
    - Record write events in the event log
    - Interact with the write model
    - Write the full state to the read model (Redis)

#### Breakdown

#### Read `balance` (or `transaction`) endpoints

- `wallet_server_service` handles `http://localhost/wallet/28skwt5B8zTrs6AqBWrSgCHLcRL/balance`
- The `Reader` abstraction has `GetBalance` called
- `GetBalance` extracts `Balance` from `GetWalletState`
- `GetWalletState` invokes `GetState`
- `GetState` attempts to get JSON from Redis
    - NOTE: We leave the `wallet_server_service` process by interacting with Redis
- Data flows back up and out to the requester (if available)

#### Write `credit` (or `debit`) events

- `wallet_server_service` handles `http://localhost/wallet/28skwt5B8zTrs6AqBWrSgCHLcRL/credit`
- The `Caller` abstraction has `Credit` called
- `Credit` invokes `Call`
- `Call` creates an event and attempts a NATS `Request` (RPC) with it
    - NOTE: We leave the `wallet_server_service` process by interacting with NATS
- `wallet_writer_service` handles the event in the NATS `Request` with the `Writer` abstraction
- The `Writer` abstraction use `handle` to record the event in the event log and pass it up to the domain implementation
- The domain implementation invokes the appropriate method against the `Wallet` abstraction
- The `Wallet` abstraction accepts or rejects the method call, possibly mutating it's state
- The domain implementation extracts the `Wallet` abstractions state and updates Redis with it
    - NOTE: At this point, a reader will see the state affected by the recently written event 
