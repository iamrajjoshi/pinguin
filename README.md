# Pinguin

Pinguin is an uptime monitoring tool for your web services written in Go.

This is very much a work in progress. My main motivation for writing this is to learn Go and to build something useful.

> [!CAUTION]
> Currently, there is no frontend for Pinguin, but it is coming soon.

## Requirements

- docker
- make

## How to Install

> [!NOTE]
> This is a work in progress. Ideally, this should be a single docker compose-file or a docker image.

To start Pinguin:

1. Clone the repository

2. Install dependencies  
   This will install the dependencies listed in the [Brewfile](Brewfile).

   ```bash
   make install
   ```

3. Start the server  
   This will use the docker compose file to spin up 3 containers:

   - `db`: TimescaleDB
   - `redis`: Redis
   - `server`: Pinguin server

   ```bash
   make up
   ```

## Pinguin Architecture

Pinguin stores monitor data in [TimescaleDB](https://www.timescale.com/) (which is a Postgres extension).

Pinguin uses Redis as its primary data pipeline for monitoring. It uses a sorted set to store the next check time for each monitor and a list to store the monitors to be checked.

```mermaid
sequenceDiagram
    participant S as Scheduler
    participant R as Redis
    participant W as Worker
    participant DB as PostgreSQL
    
    Note over S,R: Monitor Creation Flow
    S->>R: Add monitor to sorted set (ZAdd)<br/>with next check time as score
    
    Note over S,R: Scheduling Flow
    loop Every Second
        S->>R: Check sorted set for due monitors<br/>(ZRangeByScore)
        R-->>S: Return due monitor IDs
        S->>R: Push monitor IDs to work queue<br/>(LPush)
        S->>R: Reschedule monitor with new time<br/>(ZAdd)
    end

    Note over R,DB: Worker Processing Flow
    loop Continuous
        W->>R: Pop monitor from queue<br/>(BRPop)
        R-->>W: Return monitor ID
        W->>DB: Get monitor details
        DB-->>W: Return monitor config
        W->>W: Perform HTTP check
        W->>DB: Save check results
    end
```
