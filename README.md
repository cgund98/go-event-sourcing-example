# go-event-sourcing-example

This repository demonstrates how to build an **event‑sourced system** in Go.

The example application is a simplified **order management system**, inspired by platforms like eBay. It supports:

1. Tracking payment status  
2. Monitoring shipping updates  
3. Canceling orders when needed  

---

### Why Event Sourcing?

Event sourcing is an architectural pattern that stores every change to an application’s state as a sequence of events.  
It’s particularly well‑suited for systems that require:

- **Reliability** – every state change is recorded and recoverable  
- **Transparency** – a complete audit trail is available  
- **Scalability** – event logs can be processed in parallel and replayed  

For an order system, this approach provides a **robust foundation** for audits, reconciliation, and future feature growth.

---

## 🏗 Architecture Overview

This project combines **event sourcing** with **CQRS (Command Query Responsibility Segregation)**.

### Components
1. **Postgres** – Serves as both the **event store** and the **projection database**.  
2. **Kafka** – Acts as the **message bus** for notifying consumers when new events are committed.

### Why Postgres as the event store instead of Kafka?  
Kafka excels at streaming, but it’s not ideal for **per‑aggregate queries**. In this system, we frequently need to load **all events for a given `OrderID`** to rehydrate state or validate commands.  

Doing this directly from Kafka would require either:  
- Scanning the entire topic (slow and inefficient), or  
- Creating a **partition per Order** (operationally infeasible at scale).  

By writing events to **Postgres** first, we get:  
- Fast, indexed queries for any aggregate.  
- Transactional guarantees that events and downstream notifications are consistent.

### High‑level flow
1. **Commands** (e.g. `PlaceOrder`, `CancelOrder`) write new events to a Postgres **event table**.  
2. In the **same transaction**, a **notification** is published to a Kafka topic.  
3. **Consumers** read from Kafka and update **projection models** (read‑optimized views).  
4. **Projections** are stored back in Postgres and exposed to clients via **gRPC** and **JSON HTTP APIs**.

### Event Syntax

We're using **Protocol Buffers (protobuf)** to define and manage our event schemas, as they provide several unique benefits:

1. **Language‑agnostic** – We can compile stubs for multiple client languages. For example, if we later introduce a Python consumer, it can share the same schema definitions.
2. **Strong type support** – Ensures our events are well‑structured and type‑safe.
3. **Schema evolution support** – Combined with a schema registry, protobuf allows us to evolve event types in a **backwards‑compatible** way.

That said, protobuf comes with a few trade‑offs:

1. **Limited third‑party tooling** – Many analytics or BI tools don’t natively understand protobuf, meaning we’d need to serialize to JSON (or another format) for querying in a data warehouse.
2. **Not human‑readable** – The raw protobuf binary format isn’t easily interpreted in systems like Postgres; events must be parsed to make sense of their contents.
