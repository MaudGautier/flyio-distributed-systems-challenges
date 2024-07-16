# Analyzing the results

With the command used (`$MAELSTROM test -w broadcast --bin bin --node-count 5 --time-limit 20 --rate 10`), messages are
broadcast by Maelstrom to 5 nodes during 20 seconds, and they are connected with a grid topology.
Here is what a typical grid and the connection between the 5 nodes look like:

```
 n0 <-----> n1 <-----> n2
 ^          ^ 
 |          | 
 v          v
 n3 <-----> n4
```

## Expectations VS observations when broadcasting to all neighbors

In my first version, where messages received by one node are broadcast to all its neighbors, the total number of
inter-server messages expected is 10: each node will send one message to each of its neighbors:

- `n0`: 2 messages (to `n1` and `n3`)
- `n1`: 3 messages (to `n0`, `n2` and `n4`)
- `n2`: 1 message (to `n1`)
- `n3`: 2 messages (to `n0` and `n4`)
- `n4`: 2 messages (to `n1` and `n3`)

The results in that case are:

```
 :net {:all {:send-count 2252,
             :recv-count 2252,
             :msg-count 2252,
             :msgs-per-op 11.608248},
       :clients {:send-count 408, :recv-count 408, :msg-count 408},
       :servers {:send-count 1844,
                 :recv-count 1844,
                 :msg-count 1844,
                 :msgs-per-op 9.505155},
       :valid? true},
```

We get 9.5 inter-server messages, which is close to the expectations.

Note that it is slightly less than the expectations though: I will explain why in the next section.

(NB: THe real number is actually the double, since all messages get an "ok" reply. The number of messages per operation
reported in Maelstrom is also roughly divided by 2 since an "operation" counts both reads and broadcasts, but reads do
not lead to any "ok" reply between servers, only to the Maelstrom client).

## Expectations VS observations when broadcasting to all neighbors but the sender

In a second version, I decided _not_ to broadcast the message to the one neighbor from which the message originates.
That is: if `n0` broadcasts to `n1`, `n1` will itself broadcast only to `n2` and `n4` but not to `n0` since it is the
node from which it received it in the first place.

In that case, the expectations are to get 6 inter-server messages (12 if the "ok" replies were counted).

- `n0`: 1 message (to `n1` if it received it from `n3` and vice-versa)
- `n1`: 2 messages (two among `n0`, `n2` and `n4`)
- `n2`: 0 message (if it received it from `n1`)
- `n3`: 1 message (to `n0` if it received it from `n4` and vice-versa)
- `n4`: 1 message (to `n1` if it received it from `n3` and vice-versa)
  Since one of those nodes has received the message from the Maelstrom, the initial node that received it sends it to
  _all_ its neighbors (instead of all minus one). Therefore, there is one more message to add up, thus totalling to 6
  messages in total.

In practice, we get:

```
 :net {:all {:send-count 1784,
             :recv-count 1784,
             :msg-count 1784,
             :msgs-per-op 8.454976},
       :clients {:send-count 442, :recv-count 442, :msg-count 442},
       :servers {:send-count 1342,
                 :recv-count 1342,
                 :msg-count 1342,
                 :msgs-per-op 6.3601894},
       :valid? true},
```

So we get only 6.36 inter-server messages per operation. That is close to the expected 6 but there is still a small gap.
Where does it come from?

To understand, I added an extra print statement when handling a `broadcast`
message: `fmt.Fprintf(os.Stderr, "Message:%.0f,time=%s,orig=%s,dest=%s\n", body["message"], time.Now(), msg.Src, msg.Dest)`.
When looking at message 0, I get:

```
$ grep "Message:0," ./store/latest/node-logs/n*.log | cut -d":" -f2- | sort -n

Message:0,time=2024-07-13 16:32:05.616828 -0600 MDT m=+0.701117167,orig=c10,dest=n0
Message:0,time=2024-07-13 16:32:05.618824 -0600 MDT m=+0.703092168,orig=n0,dest=n3
Message:0,time=2024-07-13 16:32:05.620567 -0600 MDT m=+0.704883543,orig=n0,dest=n1
Message:0,time=2024-07-13 16:32:05.622063 -0600 MDT m=+0.706350959,orig=n3,dest=n4
Message:0,time=2024-07-13 16:32:05.623351 -0600 MDT m=+0.707638043,orig=n1,dest=n4
Message:0,time=2024-07-13 16:32:05.62429 -0600 MDT m=+0.708594584,orig=n1,dest=n2
Message:0,time=2024-07-13 16:32:05.624808 -0600 MDT m=+0.709124335,orig=n4,dest=n1
```

Here is what happens:

- The message has been sent to node `n0` by the Maelstrom client
- `n0` rebroadcasts to its two neighbors `n3` and `n1`
- `n3` rebroadcasts to its neighbor `n4` (it ignores `n0` since the message originates from it)
- `n1` rebroadcasts to its neighbors `n4` and `n2` (it ignores `n0` since the message originates from it)
- `n4` rebroadcasts to its neighbor `n1`

Let's note that `n4` has received the message from two sources (first `n3`, then `n1`), but it rebroadcasts only once:
the one originating from `n3` is rebroadcast to `n1`, but the one originating from `n1` is _not_ rebroadcast.
This is because there is an early return to avoid processing the message if it has already been seen.
All in all, there are 6 inter-server messages for broadcasting this message to all nodes, as expected.

However, in a few cases, we get the message broadcast not 6 times but 8 times between nodes:

```
$ for i in {0..83} ; do grep "Message:$i," ./store/latest/node-logs/n*.log | cut -d":" -f2- | sort -n | wl ; done | uniq -c
  83 7
   1 9

$ for i in {0..83} ; do WL=`grep "Message:$i," ./store/latest/node-logs/n*.log | cut -d":" -f2- | sort -n | wl`; if [ $WL -ne 7 ]; then echo $i ; fi ; done
59
```

There is one message in this case, and it is message #59.
When looking more closely at this message, we get:

```
$ grep "Message:59," ./store/latest/node-logs/n*.log | cut -d":" -f2- | sort -n
Message:59,time=2024-07-13 16:32:20.059177 -0600 MDT m=+15.143400793,orig=c13,dest=n3
Message:59,time=2024-07-13 16:32:20.059875 -0600 MDT m=+15.144119959,orig=n3,dest=n0
Message:59,time=2024-07-13 16:32:20.06003 -0600 MDT m=+15.144273751,orig=n3,dest=n4
Message:59,time=2024-07-13 16:32:20.060556 -0600 MDT m=+15.144828126,orig=n4,dest=n1
Message:59,time=2024-07-13 16:32:20.060557 -0600 MDT m=+15.144828918,orig=n0,dest=n1
Message:59,time=2024-07-13 16:32:20.061032 -0600 MDT m=+15.145275751,orig=n1,dest=n4
Message:59,time=2024-07-13 16:32:20.061171 -0600 MDT m=+15.145431084,orig=n1,dest=n2
Message:59,time=2024-07-13 16:32:20.061246 -0600 MDT m=+15.145505876,orig=n1,dest=n2
Message:59,time=2024-07-13 16:32:20.061912 -0600 MDT m=+15.146157167,orig=n1,dest=n0
```

Here is what happens:

- The message has been sent to node `n3` by the Maelstrom client
- `n3` rebroadcasts to its two neighbors `n0` and `n4`
- `n4` rebroadcasts to its neighbor `n1` (it ignores `n3` since the message originates from it)
- `n0` rebroadcasts to its neighbor `n1` (it ignores `n3` since the message originates from it)
- `n1` rebroadcasts to its neighbors `n4` and `n2` (it ignores `n0` since the message originates from it)
- `n1` rebroadcasts to its neighbors `n2` and `n0` (it ignores `n4` since the message originates from it)

In other words, `n1` rebroadcasts twice to its neighbors: once when it was received by `n0` and once when it was
received by `n4`.
This is likely because of a race condition: `n1` processes both messages #59 received by `n0` and by `n4` at the same
time and thus rebroadcasts both times.
This can account for the slightly higher value than expected.

## Effect of the read/broadcast ratio on the observed number of messages per operation

I observed that the number of messages per operation varies depending on the Maelstrom run.
It seems to be related to the ratio of reads over broadcast messages sent by Maelstrom.
Here are the numbers I observed on a few runs:

| # Broadcasts | # Reads | R/B Ratio | # Msgs per ops | # Msgs per broadcast (without "ok") |
|--------------|---------|-----------|----------------|-------------------------------------|
| 104          | 89      | 0.86      | 6.4974093      | 6.029                               |
| 111          | 100     | 0.91      | 6.3601894      | 6.045                               |
| 97           | 98      | 1.01      | 5.9692307      | 6.000                               |
| 102          | 107     | 1.05      | 5.866029       | 6.010                               |
| 93           | 102     | 1.10      | 5.7755103      | 6.055                               |
| 94           | 105     | 1.12      | 5.708543       | 6.043                               |
| 96           | 113     | 1.18      | 5.521531       | 6.010                               |
| 91           | 115     | 1.26      | 5.300971       | 6.000                               |
| 86           | 109     | 1.27      | 5.302564       | 6.012                               |

The higher the reads/broadcasts ratio, the lower the number of messages per operation.
Since reads lead to no inter-server messages whereas broadcasts do, it makes sense that a higher proportion of reads
will bias the number of messages per operations to be smaller.

So, the number of messages per operation should be taken as an approximation only, and it seems to be alright if the
observed number is slightly different from that expected.

A better metric would be to compute the number of messages per broadcast.
This can be done with this formula:
`#msgs_per_ops / #broadcasts * (#broadcasts + #reads)/2` (NB: I divide by 2 to remove the "ok" replies from the counts).
I added the results in the last column.
With this metric, the number of messages per broadcasts is exactly what is expected (sometimes very slightly higher,
because of the race conditions mentioned above).

## Conclusions

- 10 messages expected (and observed) when rebroadcasting to all neighbors
- 6 messages expected (and observed) when rebroadcasting to all neighbors but the sender
- Sometimes, there is a race condition that makes messages be rebroadcast more than expected (that is rare)
- The "number of messages per operation" metric is a good approximation but depends on the mix of reads and broadcasts
  done by Maelstrom. It is, however, possible to correct it to get another metric: the number of messages per broadcast.
  This one is more accurate and should be used to compare the observations and the expectations.

