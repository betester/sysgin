### A Speed Comparison between Relational vs Cache Database

This code is an experiment that i conduct on trying to know the impact of using redis and postgresSQL on handling realtime request and one of them is leaderboard. Leaderboard requires sorting on each request to get the top K ranked user.

On this instance, i create one table, which is called as leaderboard. With the following key, 

| id  | score |
| --- | ----- |
| 1   | 95.5  |
| 2   | 88.3  |
| 3   | 76.9  |

For the experiment, i don't want the network to be the bottleneck so i didn't use HTTP request, instead i use websocket. The way i conduct is similar on how online video games work, by sending concurrent request from X users which updates their score and 1 websocket connection that tells the current score. 

### Result

|Number of Request| postgreSQL  | Redis |
|---  | --- | ----- |
|10| 0ms   | 0ms  |
|100| 10ms  | 0ms  |
|1000| 500ms  | 0ms  |
|10.000| didnt try| 1ms  |

On postgesSQL it seems to hit the bottleneck when it reaches 1000 concurrent user, note that i didn't use any indexing on the score or id, which might have explained the slow performance. Regardless, using postgresSQL is much more complex if we wanted to have a faster response, compared with redis that doesn't require much configurations.

But, to make it fair i also try to optimize the query so that it uses index instead. 

### Result  : With Index

|Number of Request| postgreSQL  | Redis |
|---  | --- | ----- |
|10| 0ms   | 0ms  |
|100| 10ms  | 0ms  |
|1000| 200ms  | 0ms  |
|10.000| destroyed the processs| 1ms  |

It seems there is an improvement by using index, but not on a huge margin. The index that i use are b+tree, in the hope that finding top K element would be O(Klog(N)) where N is the number of user that tries to insert the element. But that seems still to be so slow compared with redis. On 10.000 concurrent user, there seems to be not enough port to make request to postgres and destroys my process.

### Conclusion

Redis won in terms of speed and especially matched with the curent problem, that is trying to sort fetched request. In real usecase, i doubt the number of leaderboard would reach 10.000 concurrent user, it would probably around 5 or 10 maximum. Both can be used, but it's preferable to use redis.