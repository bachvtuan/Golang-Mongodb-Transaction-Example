# Golang-Mongodb-Transaction-Example

Mongodb is good at read,write performance however it's lack of transaction unlike Sql database. However, We can fix it by using Golang we make sure read or write a row by single thread at the same time.


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


So based on this example, The maximum times user can widthdraw is 20 times ( 20 X 50$ = 1000$ ), If user can widthraw over 20 times, our system get fraud :(

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
ab -n 500 -c 100 http://localhost:8000/
```
I want to create 500 requests and there are 100 requests happend at the same time.

**Result :**

_**Unsafe code:**_
I can withdraw 500 times without any error, so actuall the money I will get is 500 * 50 = 25000USD meanwhile previous my balance is 1000USD.

_**Safe code:**_
I can withdraw 20 times, Remain requests will be shown "out_of_balance". That's what we want.


## Benmarch log

**_Unsafe code:_**
```
Server Software:        
Server Hostname:        localhost
Server Port:            8000

Document Path:          /
Document Length:        17 bytes

Concurrency Level:      100
Time taken for tests:   0.092 seconds
Complete requests:      500
Failed requests:        0
Total transferred:      67000 bytes
HTML transferred:       8500 bytes
Requests per second:    5408.21 [#/sec] (mean)
Time per request:       18.490 [ms] (mean)
Time per request:       0.185 [ms] (mean, across all concurrent requests)
Transfer rate:          707.72 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1   1.7      0       6
Processing:     3   16   8.2     15      50
Waiting:        3   16   8.2     15      50
Total:          3   17   8.6     16      52

Percentage of the requests served within a certain time (ms)
  50%     16
  66%     19
  75%     21
  80%     22
  90%     28
  95%     36
  98%     44
  99%     45
 100%     52 (longest request)
```

**_Safe code:_**

```
Server Software:        
Server Hostname:        localhost
Server Port:            8000

Document Path:          /
Document Length:        14 bytes

Concurrency Level:      100
Time taken for tests:   0.164 seconds
Complete requests:      500
Failed requests:        0
Total transferred:      65500 bytes
HTML transferred:       7000 bytes
Requests per second:    3047.67 [#/sec] (mean)
Time per request:       32.812 [ms] (mean)
Time per request:       0.328 [ms] (mean, across all concurrent requests)
Transfer rate:          389.89 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1   1.6      0       5
Processing:     0   29  13.4     29      93
Waiting:        0   29  13.4     29      93
Total:          0   30  13.6     29      97

Percentage of the requests served within a certain time (ms)
  50%     29
  66%     30
  75%     33
  80%     35
  90%     46
  95%     54
  98%     65
  99%     80
 100%     97 (longest request)
```

## References:

+ http://www.alexedwards.net/blog/understanding-mutexes
