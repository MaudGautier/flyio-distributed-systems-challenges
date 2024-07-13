# Analyzing the results on a `grid` topology

A `grid` topology corresponds to nodes connected this way:

```
n0  n1  n2  n3  n4 
n5  n6  n7  n8  n9 
n10 n11 n12 n13 n14 
n15 n16 n17 n18 n19 
n20 n21 n22 n23 n24 
```

The number of neighbors for each node is either 2 (corner nodes), 3 (border nodes) or 4 (center nodes).
For a message initially sent to a corner node (n0, n4, n20, n24), it will take 8 hops to reach all nodes (worst case is
reaching the opposite corner).
For a message initially sent to the center node (n12), it will take 4 hops to reach all nodes (worst case is reaching
corner nodes).
For a message initially sent to any other node, it will take 5 to 7 hops to reach all nodes.

Assuming that messages are initially sent with equal probability (uniform distribution) to all nodes, it takes 6.4 hops
on average for a message to reach all nodes.

To get 6.4 hops on average, we need to figure out the exact number of hops necessary to reach all nodes for each of the
25 possible starting nodes:

- 1 node (center: n12): 4 hops
- 4 nodes (center-borders: n7, n11, n13, n17): 5 hops
- 4 nodes (center-corners: n6, n8, n16, n18): 6 hops
- 4 nodes (border-middle: n2, n10, n14, n22): 6 hops
- 8 nodes (borders: n1, n3, n5, n9, n15, n19, n21, n23): 7 hops
- 4 nodes (corners: n0, n4, n20, n24): 8 hops

The weighted mean is `(1*4 + 4*5 + 4*6 + 4*6 + 8*7 + 4*8) / 25`, which is 6.4 hops on average.

## Expectations

- Max latency should be ~ 8 hops * 200 ms = 1600 ms (case where the message is broadcast to a node on the corner)
- Median latency should be ~ 6.4 hops * 200 ms = 1280 ms (see above for the explanation on the 6.4 hops on average).
- Computing the expected number of messages per ops:
    - Number of periodic broadcasts ~ (20+10) * 5 * 3.2 * 25 = 12,000 (duration: 20 seconds + 10 seconds recovery, 5
      periodic broadcasts/second, 3.2 neighbors per node on average to which periodic broadcasts are made, 25 nodes
    - Number of messages sent by Maelstrom ~ 100 * 20 = 2,000 broadcasts (rate is 100 ops per second, during 20 seconds)
    - Expected number of messages per ops ~ 12,000 / 2,000 = 6

NB: On average, each node has 3.2 neighbors to send periodic broadcasts to.
This is the weighted mean of the number of neighbors averaged over each node.
Indeed, 4 (corner) nodes have 2 neighbors, 12 (border) nodes have 3 neighbors, 9 (center) nodes have 4 neighbors.
This gives `(4*2 + 12*3 + 9*4)/25 = 3.2`

## Observations

Running the maelstrom command:

```
$ $MAELSTROM test -w broadcast --bin bin --node-count 25 --time-limit 20 --rate 100 --latency 100
```

By default, a topology `grid` is used.

### Analyzing the latency

Latency results:

```
            :stable-latencies {0 0,
                               0.5 983,
                               0.95 1341,
                               0.99 1475,
                               1 1592},
```

#### Max latency

Max latency corresponds exactly to the expectations (1592 when max 1600ms expected).

#### Median latency

Median latency is slightly smaller than the expectations (983ms when expected 1280ms).

Exactly as for the analysis on the topology `line`, this seems to be due to the fact that reads are made only every
500ms on each node.
Indeed, increasing the rate from 100 to 1000 ops per second gives a more accurate result of 1088ms, which is closer to
the expectations. There might be something else accounting for this discrepancy.

### Analyzing the number of inter-server messages per ops

Network results:

```
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
```

We observe 7 inter-server messages per operation.
As per the expectations, 6 were expected.
I do not know where the extra one comes from, but that is overall close to the expectations.

## Conclusions

The maximum latency corresponds exactly to that expected for 8 hops (i.e. a message broadcast at one corner).

The median latency is smaller than the expectation of 6.4 hops on average, but that can be accounted for by the fact
that nodes are sampled only twice per second (when increasing the rate of operations, we get results much closer to the
expectations).

The number of messages per ops corresponds roughly to the expectations.

