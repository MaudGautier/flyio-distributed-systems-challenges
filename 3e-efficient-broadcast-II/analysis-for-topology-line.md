# Analyzing the results on a `line` topology

A `line` topology corresponds to nodes connected this way:

```
n0  <---->  n1  <---->  ...  <---->  n23  <---->  n24
```

In other words, each node has 2 neighbors.
For a message initially sent to n0, it will take 24 hops to reach all nodes.
For a message initially sent to n12, it will take 12 hops to reach all nodes (bi-directionality).
For a message initially sent to n18, it will take 18 hops to reach all nodes (18 to reach n0, but only 6 to reach n24).

## Expectations

- Max latency should be ~ 24 hops * 200 ms = 4800 ms (case where the message is broadcast to a node on the edge)
- Median latency should be ~ 18 hops * 200 ms = 3600 ms (18 hops because the best case is to broadcast to the middle
  node: in that case there are 12 hops on each side, worst case is one on the edge with 24 hops.
  On average, given a uniform distribution over all nodes, we should have 18 hops).
- Computing the expected number of messages per ops:
    - Number of periodic broadcasts ~ (20+10) * 5 * 2 * 25 = 7,500 (duration: 20 seconds + 10 seconds recovery, 5
      periodic broadcasts/second, 2 neighbors per node on average to which periodic broadcasts are made, 25 nodes)
    - Number of messages sent by Maelstrom ~ 100 * 20 = 2,000 broadcasts (rate is 100 ops per second, during 20 seconds)
    - Expected number of messages per ops ~ 7,500 / 2,000 = 3.75

## Observations

Running the maelstrom command:

```
$ $MAELSTROM test -w broadcast --bin bin --node-count 25 --time-limit 20 --rate 100 --latency 100 --topology line
```

### Analyzing the latency

Latency results:

```
            :stable-latencies {0 0,
                               0.5 3051,
                               0.95 4397,
                               0.99 4617,
                               1 4781},
            :attempt-count 860,
```

#### Max latency

Max latency corresponds exactly to the expectations (4781ms when max 4800ms expected).

To analyze further the message with the worst latency, I added those logs:

- Before appending to the `messages` list upon receiving a `broadcast` message:
  `fmt.Fprintf(os.Stderr, "Adding message ?%.0f,%s,%s,%s,%s INITIAL\n", body["message"].(interface{}), time.Now(), n.ID(), msg.Dest, msg.Src)`
- Before appending to the `messages` list upon receiving a `periodic_broadcast` message:
  `fmt.Fprintf(os.Stderr, "Adding message ?%.0f,%s,%s,%s,%s \n", message, time.Now(), n.ID(), msg.Dest, msg.Src)`

Analyzing the worst (max) latency:

```
# Finding the element (message) with the worst latency:
$ cat store/latest/jepsen.log | grep -A 15 ":workload {:worst-stale"
 :workload {:worst-stale ({:element 355,
                           :outcome :stable,
                           :stable-latency 4781,

# Analyzing the hops of the longest message
$ grep "Adding message ?355," store/latest/node-logs/n*.log | cut -d":" -f2- | sort
Adding message ?355,2024-07-11 17:53:11.148764 -0600 MDT m=+8.304827084,n0,n0,c50 INITIAL
Adding message ?355,2024-07-11 17:53:11.355252 -0600 MDT m=+8.512865084,n1,n1,n0
Adding message ?355,2024-07-11 17:53:11.551029 -0600 MDT m=+8.707220043,n2,n2,n1
Adding message ?355,2024-07-11 17:53:11.750647 -0600 MDT m=+8.908342917,n3,n3,n2
Adding message ?355,2024-07-11 17:53:11.954333 -0600 MDT m=+9.111925876,n4,n4,n3
Adding message ?355,2024-07-11 17:53:12.149339 -0600 MDT m=+9.305841126,n5,n5,n4
Adding message ?355,2024-07-11 17:53:12.355244 -0600 MDT m=+9.512732501,n6,n6,n5
Adding message ?355,2024-07-11 17:53:12.555746 -0600 MDT m=+9.709325334,n7,n7,n6
Adding message ?355,2024-07-11 17:53:12.750656 -0600 MDT m=+9.905550792,n8,n8,n7
Adding message ?355,2024-07-11 17:53:12.952844 -0600 MDT m=+10.110349418,n9,n9,n8
Adding message ?355,2024-07-11 17:53:13.151398 -0600 MDT m=+10.308903126,n10,n10,n9
Adding message ?355,2024-07-11 17:53:13.350425 -0600 MDT m=+10.507725626,n11,n11,n10
Adding message ?355,2024-07-11 17:53:13.554329 -0600 MDT m=+10.711525585,n12,n12,n11
Adding message ?355,2024-07-11 17:53:13.75026 -0600 MDT m=+10.907467042,n13,n13,n12
Adding message ?355,2024-07-11 17:53:13.948305 -0600 MDT m=+11.105816126,n14,n14,n13
Adding message ?355,2024-07-11 17:53:14.154272 -0600 MDT m=+11.311963626,n15,n15,n14
Adding message ?355,2024-07-11 17:53:14.353291 -0600 MDT m=+11.510807751,n16,n16,n15
Adding message ?355,2024-07-11 17:53:14.552165 -0600 MDT m=+11.709513418,n17,n17,n16
Adding message ?355,2024-07-11 17:53:14.766545 -0600 MDT m=+11.921996792,n18,n18,n17
Adding message ?355,2024-07-11 17:53:14.953062 -0600 MDT m=+12.109896751,n19,n19,n18
Adding message ?355,2024-07-11 17:53:15.147275 -0600 MDT m=+12.304648376,n20,n20,n19
Adding message ?355,2024-07-11 17:53:15.354658 -0600 MDT m=+12.512099501,n21,n21,n20
Adding message ?355,2024-07-11 17:53:15.557927 -0600 MDT m=+12.715112042,n22,n22,n21
Adding message ?355,2024-07-11 17:53:15.751364 -0600 MDT m=+12.906220876,n23,n23,n22
Adding message ?355,2024-07-11 17:53:15.948944 -0600 MDT m=+13.102555960,n24,n24,n23
```

Observations: message sent in order from n0 to n24 with a 200ms timelapse between each hop (as expected).

#### Median latency

Median latency is slightly smaller than the expectations (3051ms when expected 3600ms).
That would be 3 fewer hops on average.

I wondered if that could be due to the fact that Maelstrom distributes more messages to the middle nodes than the
ones on the edges.

To get the distribution of nodes that receive `broadcast` messages initially, I ran:

```
$ grep "Adding message" store/latest/node-logs/n*.log | grep INITIAL | cut -d"," -f4 | sort | uniq -c | sort -n
  27 n9
  28 n21
  30 n6
  31 n17
  31 n24
  32 n20
  32 n22
  32 n8
  33 n10
  33 n14
  33 n15
  33 n19
  33 n3
  33 n4
  34 n5
  35 n12
  36 n1
  36 n11
  36 n2
  37 n13
  37 n18
  37 n7
  38 n23
  39 n16
  54 n0
```

So all nodes receive approximately 33 messages, except for node n0 which receives more than 50.
(I don't know why n0 is a special node but that is consistent when reproducing multiple times.
In any case, this discrepancy is very small and does not matter when averaged over the 25 nodes.)

Using these numbers, we can get the total number of hops necessary for all messages to reach its furthest node (if it is
node n12, then the number of hops is 12, for all others it is 12 + the distance to n12, or else `max(24-X, X)` where `X`
is the node number `nX`). That gives:
`27*(24-9)+28*21+30*(24-6)+31*17+31*(24)+32*20+32*22+32*(24-8)+33*(24-10)+33*14+33*15+33*19+33*(24-3)+33*(24-4)
+34*(24-5)+35*12+36*(24-1)+36*(24-11)+36*(24-2)+37*13+37*18+37*(24-7)+38*23+39*16+54*(24-0)`
which sums up to 15783 hops. Since each needs 200ms, and we have 860 messages (`attempt-count` in the results),
that would give `15783*200/860 = 3670ms` as the average latency.
This value corresponds exactly to the expected median latency.

So far, I don't understand why the observed median latency is smaller.
Part of the reason would be that rate is 100 operations per second.
Given that half are read, that gives 50 reads per second, or 2 reads per node per second (25 nodes).
Since we can sample reads on each node only once every 500ms, that could account for the difference.

I tested with a rate of 1000 operations per second to see if I got a better estimation for the median latency:

```
$MAELSTROM test -w broadcast --bin bin --node-count 25 --time-limit 20 --rate 1000 --latency 100 --topology line
            :stable-latencies {0 0,
                               0.5 3436,
                               0.95 4644,
                               0.99 4831,
                               1 5491},
            :attempt-count 3812,
```

The new median latency is indeed 3436, thus much closer to the expectations (I also get one message at 5491ms, but that
could be due to overload and wait time to be processed since messages are much more frequent and there may be too much
traffic to process).

### Analyzing the number of inter-server messages per ops

Network results:

```
 :net {:all {:send-count 10690,
             :recv-count 10690,
             :msg-count 10690,
             :msgs-per-op 6.321703},
       :clients {:send-count 3482, :recv-count 3482, :msg-count 3482},
       :servers {:send-count 7208,
                 :recv-count 7208,
                 :msg-count 7208,
                 :msgs-per-op 4.2625666},
       :valid? true},
```

We observe 4.26 inter-server messages per operation.
This is very close to the expected 3.75 messages per ops.

For more explanations:
Each node does about 300 periodic broadcasts:

- Periodic broadcasts are made during 30 seconds (20 time-limit + 10 recovery)
- There are 5 periodic broadcasts per second (every 200ms)
- Each periodic broadcast is sent to every neighbor, and on a line topology, all nodes have 2 neighbors (except the 2 on
  the edges which have only 1 each)
- There are 25 nodes
- All in all: 30 * 5 * 2 * 25 = 7,500 periodic broadcasts over the experiment.

The total number of operations sent by Maelstrom is: 100 * 20 (rate: 100 ops/second, time-limit: 20 seconds), i.e. 2,000
operations.

So the expected number of messages per operation is 7,500/2,000 = 3.75, not far from the 4.26 observed.

## Conclusions

The maximum latency corresponds exactly to that expected for 24 hops (i.e. a message broadcast at one edge).

The median latency is smaller than the expectation of 18 hops on average, but that can be accounted for by the fact that
nodes are sampled only twice per second (when increasing the rate of operations, we get results much closer to the
expectations).

The number of messages per ops corresponds to the expectations
