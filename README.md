
# Go RabbitMQ

This project provides a simple example of using RabbitMQ in Go.

It demonstrates how to implement:

* **Direct exchange**
* **Fanout exchange**
* **Topic exchange**
* **Headers exchange**
* **Dead Letter Exchanges (DLX)**
* Basic Publisher/Consumer pattern using RabbitMQ
* CLI interface using Cobra

## 🐇 Basic Concepts of RabbitMQ

RabbitMQ is a **message broker** that allows applications to communicate by sending and receiving messages through queues.

Key concepts:

### 1️⃣ Producer

* An application that sends messages.

### 2️⃣ Consumer

* An application that receives messages from a queue.

### 3️⃣ Queue

* A buffer that stores messages.
* Consumers subscribe to a queue to receive messages.

### 4️⃣ Exchange

* Receives messages from producers and routes them to queues based on rules.
* Types of exchanges:

  * **Direct** — Routes messages to queues based on exact routing key.
  * **Fanout** — Broadcasts messages to all bound queues.
  * **Topic** — Routes messages based on pattern matching in the routing key.
  * **Headers** — Routes messages based on header attributes.

---

| Exchange Type        | Use Case / Example |
| -------------------- | ------------------ |
| **Direct Exchange**  | Send specific notifications to a particular system. <br>🛒 `order.shipped` → shipping system. <br>💳 `order.canceled` → refund system. |
| **Fanout Exchange**  | Broadcast messages to all subscribers. <br>📢 Maintenance announcement → all mobile apps, email service, admin dashboard. |
| **Topic Exchange**   | Flexible message routing using patterns. <br>📦 `transaction.purchase.electronics` → electronics consumer. <br>↩️ `transaction.return.*` → all return handling systems. |
| **Headers Exchange** | Routing based on header key/value. <br>📄 Report `type=sales` + `format=pdf` → PDF report queue. <br>📊 Report `type=inventory` + `format=excel` → Excel report queue. |

---

| Exchange Type        | Usage Frequency | Typical Use Case |
| -------------------- | ----------------| ---------------- |
| **Direct Exchange**  | ⭐⭐⭐⭐⭐ (very common) | Simple routing with exact routing key. Suitable for clear business events: order.created, user.updated, etc. |
| **Topic Exchange**   | ⭐⭐⭐⭐ (common) | Flexible and scalable routing, suitable for event-driven systems/microservices. Supports wildcard (`*`, `#`). |
| **Fanout Exchange**  | ⭐⭐ (occasional) | Broadcast use cases, used in notifications where all consumers must receive the message (e.g., push notifications, broadcast events). |
| **Headers Exchange** | ⭐ (rare) | Special cases where header metadata strictly determines routing. Often replaced with Topic + manual header processing. |

---

### 🎭 Wildcard in **Topic Exchange**

* Routing key in Topic Exchange typically follows **dot-separated string** format, for example:

  ```
  order.created
  order.updated.customer
  transaction.purchase.electronics
  ```

* Binding key (used when binding a queue to a topic exchange) can use **wildcards**:

| Wildcard | Meaning |
| -------- | ------- |
| `*`      | **One word** (one segment between dots `.`). |
| `#`      | **Zero or more words** (can be empty, one word, or multiple words). |

---

### 🔎 Usage Examples

#### 1️⃣ `*` → **exactly one word**

```text
Binding key: order.*

Matches:
✅ order.created
✅ order.canceled

Does not match:
❌ order.updated.customer (because it has 2 words after order)
```

#### 2️⃣ `#` → **zero or more words**

```text
Binding key: order.#

Matches:
✅ order.created
✅ order.canceled
✅ order.updated.customer
✅ order.updated.customer.address
✅ order

Why does it match `order`? Because `#` can match "zero words".
```

---

### Practical Summary

| Wildcard | When to Use |
| -------- | ----------- |
| `*`      | When you want to match **only one specific segment**. |
| `#`      | When you want to match **anything after a prefix**, or the entire routing key. |

---

### Analogy

* `*` is like **match one level** → `order.*` → everything one level after `order`.
* `#` is like **match recursively** → `order.#` → everything that starts with `order`, regardless of the level depth.

### 5️⃣ Binding

* A rule that connects an exchange to a queue.

### 6️⃣ Routing Key

* A key used by Direct and Topic exchanges to decide how to route the message.

### 7️⃣ Dead Letter Exchange (DLX)

* A special exchange used to handle messages that cannot be delivered or were rejected.

### Flow Example

```
  Producer
    |
  Exchange
    |
  [Binding] -- (routing key / headers)
    |
  Queue
    |
  Consumer
```

### Exchange Types Diagram

```
Direct: routingKey="order.created"
Topic:  routingKey="order.*"
Fanout: broadcast to all queues
Headers: match headers {"type": "report"}
```

### Why RabbitMQ?

* Decouple services
* Asynchronous communication
* Scalable and resilient architecture
* Retry and Dead Lettering support

## Structure

```
go-rabbitmq/
├── cmd/                # CLI commands using Cobra
│   ├── listener.go     # CLI for starting listeners
│   ├── publisher.go    # CLI for publishing messages
│   └── root.go         # CLI root command
├── consumer/           # Example consumer
├── producer/           # Example publisher
├── pkg/rabbitmq/       # RabbitMQ abstraction layer
│   ├── client.go
│   ├── listener.go
│   └── publisher.go
├── main.go             # Entry point
├── go.mod
└── README.md
```

## Usage

### 1️⃣ Build the project

```bash
go build -o go-rabbitmq
```

### 2️⃣ Run CLI commands

#### Publish message

```bash
./go-rabbitmq publish --exchange-type=direct
./go-rabbitmq publish --exchange-type=fanout
./go-rabbitmq publish --exchange-type=topic
./go-rabbitmq publish --exchange-type=headers
```

#### Run listener

```bash
./go-rabbitmq listen
```

### 3️⃣ Example message payload

#### Direct / Topic

```json
{
  "orderId": "12345",
  "status": "shipped"
}
```

#### Fanout

```json
{
  "event": "system.broadcast",
  "message": "Hello all subscribers!"
}
```

#### Headers

```json
{
  "reportId": "98765",
  "content": "PDF report content"
}
```

Headers:

```yaml
type: report
format: pdf
```

### 4️⃣ Example DLX (Dead Letter Exchange)

* If your consumer returns an error, the message can be routed to a **DLX queue**.
* DLX queues are declared in `RunListener()` with:

  * `x-dead-letter-exchange`
  * `x-dead-letter-routing-key`

Example DLX message log:

```txt
[DLX] Received dead letter message: {...}
```

## Configuration

RabbitMQ connection URL (default):

```
amqp://guest:guest@localhost:5672/
```

## Requirements

* Go >= 1.20
* RabbitMQ server running

## Notes

* This project uses a modular design and abstracts RabbitMQ details via `pkg/rabbitmq`.
* Useful for learning or quick-start RabbitMQ integration in Go projects.
* Based on `amqp091` RabbitMQ client for Go.
* CLI based on `cobra` library.

## License

MIT
