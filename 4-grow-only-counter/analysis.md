# Analysis

## Sequential consistency

Sequential consistency guarantees that all operations appear to take place in some total order, and that the order is
consistent with the order of operations on each individual process.

In practice, one way to implement sequential consistency is to use a main process that consists of a queue.
Writes from all other processes are sent to that queue and thus take place in some total order (and in an order coherent
with the order of those operations on each process).
These writes are then applied in the order of the queue to all other processes, guaranteeing the same order for all
processes.
As for reads, they can be done locally (i.e. on each process).
Reads can thus be stale (if some write operations have not been made on the process), but they are guaranteed to be in a
meaningful order (per process).

## Implementation choices

- In the sequentially-consistent KV store: Initialize a global counter
- `add`: read the current value of the global counter, add up the new value and update the global counter with a
  "compare-and-swap" operation (retry until it works) => guarantees that the new value has not skipped any writes: we
  have read the last value, and record locally (in the node) the value written
- `read`: ask all nodes for the last value they wrote and take the most up to date (max) of those
- `last_written`: to reply to another node's request to have the last value written

Basically, the solution to deal with a sequentially consistent store, when each process is sticky connected to one node
that may have stale data, is to have all nodes talk to one another to find the most up-to-date information before
serving the client.

## Resources

- Sequential consistency explained by Martin Kleppmann during a "Papers we love"
  meetup: https://martin.kleppmann.com/2015/07/08/attiya-welch-at-papers-we-love.html

