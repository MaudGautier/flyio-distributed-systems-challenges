# Fly.io distributed systems challenges

This repo contains my solutions to [Gossip Glomers](https://fly.io/dist-sys/), a series of distributed systems
challenges available made by [Fly.io](https://fly.io/).

The challenges are built on top of [Maelstrom](https://github.com/jepsen-io/maelstrom), a platform that lets us build a
representation of a node in a distributed system and handles the routing of messages between those nodes.

## Getting started

Download and install the maelstrom package by downloading the latest
tarball ([latest as of now: 0.2.3](https://github.com/jepsen-io/maelstrom/releases/tag/v0.2.3)).

```shell
# Set the Maelstrom path to its executable
export MAELSTROM=/path/to/maelstrom

# Move into any challenge directory and execute the test file (that build the binary and executes it) 
cd challenge-X
bash test.sh
```

## Challenge 1: Echo

**Goals:**

- Get started with Maelstrom and message specifications
- Build a simple handler to reply to `echo` messages received

[Solution](./1-echo/main.go)

**Understandings:**

- A message sent by the Maelstrom client is always of this form:

```
{
  "src": "c1",
  "dest": "n1",
  "body": {
    "type": "echo",
    "msg_id": 1,
    "other_field": "Other"
  }
}
```

- To debug: print errors to stderr (with `fmt.Fprintf(os.Stderr, "Printing %s \n", variable)`). Otherwise, Maelstrom
  crashes!
- Even if there is a crash at the beginning, the Maelstrom program runs until the specified time (set
  via `--time-limit`). The potential errors can be searched in `challenge-X/store/latest/node-logs/nX.log`.
- Some statistics are given at the end in this form:

```
{:perf {:latency-graph {:valid? true},
        :rate-graph {:valid? true},
        :valid? true},
 :timeline {:valid? true},
 :exceptions {:valid? true},
 :stats {:valid? true,
         :count 44,
         :ok-count 44,
         :fail-count 0,
         :info-count 0,
         :by-f {:echo {:valid? true,
                       :count 44,
                       :ok-count 44,
                       :fail-count 0,
                       :info-count 0}}},
 :availability {:valid? true, :ok-fraction 1.0},
 :net {:all {:send-count 90,
             :recv-count 90,
             :msg-count 90,
             :msgs-per-op 2.0454545},
       :clients {:send-count 90, :recv-count 90, :msg-count 90},
       :servers {:send-count 0,
                 :recv-count 0,
                 :msg-count 0,
                 :msgs-per-op 0.0},
       :valid? true},
 :workload {:valid? true, :errors ()},
 :valid? true}


Everything looks good! ヽ(‘ー`)ノ
```

## Challenge 2: Unique ID generation

**Goals:**

- Implement a globally-unique ID generation system
- It should be totally available, i.e. continue to operate in case of network partitions

[Solution](./2-unique-ids/main.go)

I concatenated the node's ID with a counter for each node. The unique IDs
are: `n0-0`, `n0-1`, `n0-100...`, ...., `n1-0`, ..., `n2-100...`.

More explanations:

- Each node must generate new unique IDs even if there is a network partition (i.e. if it cannot speak to other nodes
  for a given period of time).
  Therefore, there must be something unique to each node: I took the node's ID. (NB: this assumes that the system that
  assigns IDs to each node already works as expected, i.e. node IDs are unique at all times.)
- On top of that, I decided to make the IDs monotonically increasing, and thus used a counter to compute the second part
  of the ID.

Some other alternatives considered:

- Use a UUID as the unique ID. This would guarantee uniqueness (if we generate 1 billion UUIDs per second during 86
  years, there is a 50% probability that there is at least one collision -
  see [wikipedia page](https://en.wikipedia.org/wiki/Universally_unique_identifier))
- If the node can crash and recover (not the case in the Maelstrom experiments), the last ID used should be persisted on
  disk to avoid collisions (in the case of a monotonically increasing ID).
- If we could not trust the node ID generator, we could use the MAC address of each node (+ for e.g. a timestamp if it
  was useful for the use case).

## Challenge 3: Broadcast

### Challenge 3a: Single-node broadcast

**Goal:**
Implement the message handlers for a single-node broadcast system (first step before incrementing the full broadcast
system in which messages are gossiped from one node to the other).

[Solution](./3a-single-node-broadcast/main.go)

I added 3 handlers:

- A `broadcast` handler that appends the received message to an in-memory list of messages received by the node
- A `read` handler that returns the full list of messages that the node has received
- A `topology` handler that only replies `ok` when such a message is received (later, in challenge 3b where there are
  multiple nodes, I parse the topology and record it to broadcast messages as desired).

### Challenge 3b: Multi-node broadcast

**Goal:**
Propagate messages to other nodes in the cluster

[Solution](./3b-multi-node-broadcast/main.go)

My solution:

- Updated the `topology` handler to take the topology into account (find out which nodes are neighbors/can communicate
  with one another)
- Upon receiving a `broadcast` message:
    - Check if it is already in the list of seen messages. If yes, do nothing (early return).
    - Rebroadcast it to all the nodes' neighbors (based on the topology) except for the sender node (which sent the
      message to the current node in the first place and thus already has it)

[Detailed analysis of the metrics observed](./3b-multi-node-broadcast/analysis.md)

TL;DR of the analysis:

- Based on the topology (grid) and the number of nodes (5), 6 inter-server messages are expected per broadcast.
- We observe exactly 6 inter-server messages per broadcast or slightly higher.
- The slightly higher number comes from the fact that there are a few race conditions (about 1% of cases). These race
  conditions bypass the early return when the message has already been seen and thus lead to broadcasting slightly more
  messages than expected.
- The metric given by Maelstrom "number of messages per operation" is biased depending on the read/broadcast mix in its
  operations: the more reads, the lower the number (since reads do not lead to any inter-server message). Nonetheless,
  this biased metric can be corrected to obtain the "number of inter-server messages per broadcast". With this, we get
  _exactly_ the expected number of messages per broadcast. Also, it is highly reproducible and with an extremely low
  variance.

### Challenge 3c: Fault-tolerant broadcast

**Goal:**
Handle network partitions (i.e. some nodes cannot communicate with one another)

How Maelstrom proceeds:

1. Create 2 partitions of nodes
2. Broadcast message to nodes for the duration defined in the parameters (`time-limit`)
3. Stop broadcasting and heal network, i.e. nodes can communicate with one another again
4. Wait for 10 seconds and analyze the results (read all nodes to check if they all have all messages)

[Solution](./3c-fault-tolerant-broadcast/main.go)

2 solutions considered:

1. When rebroadcasting a message, wait for an `ok` response and retry sending after, say, 2 seconds if it was not
   received
2. Add a periodic broadcast where all messages seen so far are broadcast to all neighbors (say, every 2 seconds).

I opted for solution #2 because it seemed simpler to implement and because it would cause much fewer messages to be sent
over the network.
The periodic broadcast is done in a go routine, and appends each message received to the list of messages.

### Challenge 3d: Efficient broadcast, part I

**Goals:**

- Broadcast all messages to all nodes with:
    - Number of messages per operation < 30
    - Median latency < 400 ms
    - Max latency < 600 ms
- The solution should work in case of a partition (with different statistics, not specified).

[Solution](./3d-efficient-broadcast-I)

Given that, in this challenge, there is a simulated network latency of 100ms (i.e. each node will wait 100ms when
sending a message), and that the max latency should be 600ms to reach all nodes, there should be at most 6 hops to reach
the farthest node.
With a grid topology of 25 nodes, this is _not_ possible (from one corner to another, there are 8 hops).
Therefore, it is necessary to change the topology to another one that would match with this limit of 6 hops.
With a tree topology where each node has 2 or 3 children, it works (it would require log2(25) or log3(25) hops, both of
which are below 5 hops). This would seem like a realistic solution in the real world.
A simpler way to implement the solution is to use a flat tree topology, where one node speaks to all other nodes.
That is the way I implemented it.

### Challenge 3e: Efficient broadcast, part II

**Goals:**

- Broadcast all messages to all nodes with:
    - Number of messages per operation < 20
    - Median latency < 1,000 ms
    - Max latency < 2,000 ms
- The solution should work in case of a partition (with different statistics, not specified).

[Solution](./3e-efficient-broadcast-II/main.go)

There are 25 nodes, so each message should be sent to all other 24 nodes.
Since we need to have fewer than 20 inter-server messages per operation, messages cannot be sent one by one. They must
be batched.

In my implementation:

- Upon receiving a `broadcast` message from the Maelstrom client, I add the message to the node's list of messages and
  do nothing else
- Every 200ms, a `periodic_broadcast` message containing the batch of all messages that the node knows of is sent to its
  neighbors.

With a `grid` topology and this implementation, I get those results:

```
-- VALIDITY RESULTS:
Everything looks good! ヽ(‘ー`)ノ

-- MESSAGE PER OPS RESULTS:
 :net {:all {:send-count 15618,
             :recv-count 15572,
             :msg-count 15618,
             :msgs-per-op 9.085515},
       :clients {:send-count 3538, :recv-count 3538, :msg-count 3538},
       :servers {:send-count 12080,
                 :recv-count 12034,
                 :msg-count 12080,
                 :msgs-per-op 7.0273414},
       :valid? true},

-- LATENCY RESULTS:
            :stable-latencies {0 0,
                               0.5 983,
                               0.95 1341,
                               0.99 1475,
                               1 1592},

-- EXPECTATIONS:
   - msgs-per-ops   < 20
   - median latency < 1,000ms
   - max latency    < 2,000ms
```

Quick explanations of the numbers (read the [full analysis](./3e-efficient-broadcast-II/analysis-for-topology-grid.md)
for a more detailed understanding):

- The max latency is 1592 ms. This corresponds to the worst case scenario where the message is initially broadcast to a
  node in the corner. In that case, 8 hops are needed to reach all nodes, which makes `8 * 200 ms = 1600 ms` in the
  worst case.
- The median latency is 983 ms which is less than the specifications. Theoretically, this could go up to 1,280 ms, as
  6.4 hops are required on average (`6.4 * 200 = 1,280 ms`). To make sure I get below 1,000 ms every time, I should have
  set my periodic broadcast to run once every 150 ms.
- The number of inter-server messages per operation is 7. That is close to the expected 6 messages per operation (
  details in the full analysis file).

I did a much more detailed analysis of these results both for a `line` topology (much easier to grasp, since the
topology is simpler) and for the `grid` topology. Here they are:

- [Analysis for the line topology](./3e-efficient-broadcast-II/analysis-for-topology-line.md) (Very detailed, easy to
  understand, some extra analyses)
- [Analysis for the grid topology](./3e-efficient-broadcast-II/analysis-for-topology-grid.md) (More succinct but
  interesting to get the hang of how to think the grid)

## Challenge 4: Grow-only counter

**Goal:**
Implement a stateless grow-only counter that relies on a sequentially-consistent key-value store.

[Solution](./4-grow-only-counter/main.go)

A sequentially-consistent key-value store guarantees that all operations appear to take place in some total order, and
that the order is consistent with the order of operations on each individual process.

To solve the challenge, I implemented the following algorithm:

- For writes: I make sure that the new counter value includes all the previous ones by first reading the latest counter
  value in the KV store, second recomputing the new value and third doing a "compare-and-set" operation. This process is
  retried until it succeeds, which guarantees that the new write made has not removed any prior writes that could have
  been done by other nodes.
- For reads: Reading from the sequentially-consistent KV-store may give stale values. To account for that staleness, the
  node reading asks all other nodes for the last value they wrote. The most recent one (i.e. the biggest value in the
  case of the grow-only counter) is the one returned when serving the request.

A more detailed explanation of this solution is in [this analysis](./4-grow-only-counter/analysis.md)

