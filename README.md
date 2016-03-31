# Golang-Mongodb-Transaction-Example

Mongodb is good at read,write performance however it's lack of transaction unlike Sql database. However, We can fix it by using Golang we make sure read or write a row by single thread at the same time. I see they have already written the document how to handle transaction by using code [here](https://docs.mongodb.org/manual/tutorial/perform-two-phase-commits/) but it's too complicated and hard to implement so it take a lot of time too.


This is included examples to guide how to make transaction on Mongodb by using programming language layer is Golang.

## Install dependency lib

```
go get gopkg.in/mgo.v2
```
This lib handle communicate to Mongodb

## Scenario:

It demonstrate a simple server that serve user withdraw money from the bank.
**Example steps:**

1.  Init bank account with an amount is 1000USD.
2.  If there is a request is called to server, user will withdraw 50$ if current balance is not less than 50$.
3.  If widthdraw is ok, calculate remain balance then update to DB.


So based on this example, The maximum times user can widthdraw is 20 times ( 20 X 50$ = 1000$ ), If user can widthraw over 20 times, our system get fraud.

Usually, If every request is sequence, the fraud will not happend. However, if there are many requests happend at the same time ( over 20 requests ), The fraud will happend if we use basic code.

## Test log:

Test tool: I use simple test tool called ab test


*Install:*

```
sudo apt-get install apache2-utils
```

*Sytax:*

```
ab -n <num_requests> -c <concurrency> <addr>:<port><path>
```

Example usage:

```
ab -n 5000 -c 1000 http://localhost:8000/
```
I want to create 5000 requests and there are 1000 requests happend at the same time.

## Alternative method:

### Queue ( queue_code.go )
This method can be applied for other programming language. The idea is we should implement queue for it.

Each payment is processed one by one.Take a look at queue_code.go. I have implemented 2 channels. The first one is input channel and other one is output channel.

### Multiple queues ( multiple_queue_code.go )
By using above method, I use single queue, and now let expland by using multiple queue will make server response faster. To do this, I generated N users in system. Each user need belong only 1 queue by your rule.

**For example:**

I have 100 users and there are 10( Q ) queues are listening are numbered from 0 -> 9( 10 -1 ). If user X ( 0-> 99 ) want to withdraw I calculate what queue it should be used. My rule is simple by get modulo of X by  Q.

+ X = 52, Q = 10 -> The queue should be process for this request is 52 % 10 = 2 ( third queue )
+ X = 20, Q = 10 -> The queue should be process for this request is 20 % 10 = 0 ( first queue )



## Result :

_**Unsafe code( Not passed ):**_
I can withdraw 1653 times without any error, so actuall the money I will get is 1653 * 50 = 82650USD. This is serious problem because my previous balance only 1000USD.

_**Safe code( Passed ):**_
I can withdraw 20 times, Remain requests will be shown "out_of_balance". I watched the console log and see the result as expected ( Good job ).

_**Queue( Passed ):**_
Unlike above methods, we should simulate multiple users for system to test the result don't messed between users. I generated 100 accounts, each account has 1000USD. The the maxinum successful withdraw times are 2000 ( 20/user * 100 ). I watched the console log and see the result as epxectation too.

_**Multiple Queues( Passed ):**_
The same behaviour with Queue method but There are N queues are listening so the execution time is shorter than singlular queue. I believe this way is the best. Look at the time taken in the benchmark log.


## Benchmark log

**_Unsafe code:_**
```
Concurrency Level:      1000
Time taken for tests:   0.599 seconds
Complete requests:      5000
Failed requests:        3711
   (Connect: 0, Receive: 0, Length: 3711, Exceptions: 0)
Total transferred:      638243 bytes
HTML transferred:       54532 bytes
Requests per second:    8341.29 [#/sec] (mean)
Time per request:       119.886 [ms] (mean)
Time per request:       0.120 [ms] (mean, across all concurrent requests)
Transfer rate:          1039.80 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1   5.4      0      32
Processing:     0   18  12.7     16     227
Waiting:        0   18  12.7     15     227
Total:          0   19  13.9     16     251

Percentage of the requests served within a certain time (ms)
  50%     16
  66%     21
  75%     24
  80%     27
  90%     39
  95%     47
  98%     55
  99%     58
 100%    251 (longest request)
```

**_Safe code:_**

```
Concurrency Level:      1000
Time taken for tests:   1.300 seconds
Complete requests:      5000
Failed requests:        4980
   (Connect: 0, Receive: 0, Length: 4980, Exceptions: 0)
Total transferred:      654740 bytes
HTML transferred:       69760 bytes
Requests per second:    3847.42 [#/sec] (mean)
Time per request:       259.914 [ms] (mean)
Time per request:       0.260 [ms] (mean, across all concurrent requests)
Transfer rate:          492.00 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0   67 248.3      0    1001
Processing:     0   51  42.8     43     284
Waiting:        0   51  42.8     43     284
Total:          0  118 276.3     43    1285

Percentage of the requests served within a certain time (ms)
  50%     43
  66%     47
  75%     57
  80%     76
  90%     87
  95%   1065
  98%   1246
  99%   1267
 100%   1285 (longest request)
```

**_Queue method code:_**
```
Concurrency Level:      1000
Time taken for tests:   1.439 seconds
Complete requests:      5000
Failed requests:        4991
   (Connect: 0, Receive: 0, Length: 4991, Exceptions: 0)
Total transferred:      663893 bytes
HTML transferred:       78893 bytes
Requests per second:    3475.43 [#/sec] (mean)
Time per request:       287.734 [ms] (mean)
Time per request:       0.288 [ms] (mean, across all concurrent requests)
Transfer rate:          450.65 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0   79 268.1      0    1004
Processing:     4   97  33.5     95     323
Waiting:        4   97  33.5     95     323
Total:         31  177 275.4     98    1326

Percentage of the requests served within a certain time (ms)
  50%     98
  66%    114
  75%    125
  80%    126
  90%    134
  95%   1087
  98%   1132
  99%   1143
 100%   1326 (longest request)

```

**_Multiple Queue method code:_**
```
Concurrency Level:      1000
Time taken for tests:   0.604 seconds
Complete requests:      5000
Failed requests:        4991
   (Connect: 0, Receive: 0, Length: 4991, Exceptions: 0)
Total transferred:      664353 bytes
HTML transferred:       79353 bytes
Requests per second:    8281.00 [#/sec] (mean)
Time per request:       120.758 [ms] (mean)
Time per request:       0.121 [ms] (mean, across all concurrent requests)
Transfer rate:          1074.51 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1   5.0      0      31
Processing:     0   21  14.1     18     236
Waiting:        0   21  14.1     18     236
Total:          0   22  15.6     18     256

Percentage of the requests served within a certain time (ms)
  50%     18
  66%     27
  75%     32
  80%     35
  90%     42
  95%     50
  98%     60
  99%     65
 100%    256 (longest request)

```

## References:

+ http://www.alexedwards.net/blog/understanding-mutexes
