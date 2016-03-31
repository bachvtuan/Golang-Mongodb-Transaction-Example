/**
 * This is the example code demonstrate transaction about transfer money between accounts in system. 
 * Test condition: the total balance of all users before and after transfer should have the same amount.
 */
package main

import (
  "strconv"
  "io"
  "net/http"
  "fmt"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "math/rand"
  "time"
  "sync"
)

var global_db *mgo.Database

var mu = &sync.Mutex{}

//Get random number from range [ min, max ]
func Random(min, max int) int {

  rand.Seed(time.Now().UTC().UnixNano())
  return rand.Intn(max - min + 1) + min
}

type Currency struct {
  Id   bson.ObjectId `json:"id" bson:"_id,omitempty"`
  Amount float64 `bson:"amount"`
  Account   string `bson:"account"`
  Code   string `bson:"code"`
}


var countTrasaction = 0
var maxUser = 100
var maxThread = 10

//Array of channels input and output
var in []chan Transaction
var out []chan Result

type Transaction struct{
  Source string
  Target string
}

type Result struct{
  Account string
  Result string
}

func transfer(w http.ResponseWriter, r *http.Request) {
  


  // random user from 1 to maxUser
  number := Random( 1, maxUser )

  //Allocate to appropriate channel number based on number by get the last number in the random number.
  channelNumber :=  number % maxThread
  sourceAccount := "user" + strconv.Itoa( number )
  targetNumber := Random( 1, maxUser )

  if targetNumber == number{
    io.WriteString(w, "ignore because same account")
    return
  }

  targetAccount := "user" + strconv.Itoa( targetNumber )

  var wg sync.WaitGroup
  wg.Add(1)

  go func () {
    
    in[ channelNumber ] <-  Transaction{ Source: sourceAccount, Target: targetAccount }

    for {
      select {
      case result := <- out[ channelNumber ]:

        if result.Account == sourceAccount{
          /*fmt.Printf("Result %s\n", result.Result)
          fmt.Printf("Number is %d \n", channelNumber )*/
          fmt.Printf("Result %s and countTrasaction is %d\n", result.Result, countTrasaction)
          io.WriteString(w, result.Result)
          wg.Done()
          //should return, otherwise it's still pop out value from out channel
          return          
        }else{
          fmt.Printf("Dismatch: %s and %s\n", result.Account, sourceAccount)
          panic("why ?, Something went wrong")
          //push to out again
          out[ channelNumber ] <- result
        }

      };
    }    
  }()


  wg.Wait()
  
}

func main() {

  in = make([]chan Transaction, maxThread)
  out = make([]chan Result, maxThread)

  for i := range in {

    fmt.Printf("i %d \n", i )
    in[i] = make(chan Transaction)
    out[i] = make(chan Result)
  }

  

  session, _ := mgo.Dial("localhost:27017")
  fmt.Printf("Session is %p\n", session)
  global_db = session.DB( "db_log" )

  //make sure it is empty first
  global_db.C("bank").DropCollection()
  
  //Init maxUser with amount are 1000USD.
  for i := 1; i <= maxUser; i++ {
    user := Currency{ Account : "user" + strconv.Itoa( i ) , Amount: 1000.00, Code:"USD" }
    err := global_db.C("bank").Insert(&user)

    if err != nil{
      panic("insert error")
    }          
  }

  fmt.Printf("len in is %d", len( in ))
  fmt.Printf("len out is %d", len( out ))

  //Create 10 go routine to handle for each channel
  for i := range in {

    go func ( subIn *chan Transaction, index int ) {

      for {
        select{
          case transaction := <-*subIn:
              
             account := transaction.Source

            fmt.Printf("On worker %d \n", index + 1)
            /*count_queue += 1
            fmt.Printf("count_queue %d\n", count_queue)*/
            entry := Currency{}
            err := global_db.C("bank").Find(bson.M{"account":  account }).One(&entry)

            if err != nil {
              panic(err)
            }

            
            if entry.Amount < 50.00 {
              out[ index ] <-  Result{ Account: account, Result: "out_of_balance"}
            }else{
              
              // Decrease balance from source account
              /**
              * Should not use this code to update
              *     entry.Amount = entry.Amount - 50.00
              *     err = global_db.C("bank").UpdateId(entry.Id, entry)
              * Because maybe other go routine are handle this source account the target account. 
              */
              colQuerier := bson.M{"account": transaction.Source }
              change := bson.M{"$inc": bson.M{"amount": -50 }}
              err = global_db.C("bank").Update(colQuerier, change)

              if err != nil {
                out[ index ] <-  Result{ Account: account, Result: "update error"}
              }

              // Increase balance to target account
              colQuerier = bson.M{"account": transaction.Target }
              change = bson.M{"$inc": bson.M{"amount": 50 }}
              err = global_db.C("bank").Update(colQuerier, change)
              if err != nil {
                out[ index ] <-  Result{ Account: account, Result: "update error"}
              }
              
              
              countTrasaction = countTrasaction + 1
              
              fmt.Printf("countTrasaction %d\n", countTrasaction)
              

              out[ index ] <-  Result{ Account: account, Result: fmt.Sprintf("countTrasaction %d\n", countTrasaction)}
            }
        }
      }

    }(&in[i], i)
  }

  
  http.HandleFunc("/", transfer)
  http.ListenAndServe(":8000", nil)
}

/*
How to test this code works:

After init: we can use this command to show the total amount of all users.
db.getCollection('bank').aggregate(
   [
     {
       $group:
         {
           _id: null,
           totalAmount: { $sum: "$amount" },
           count: { $sum: 1 }
         }
     }
   ]
)
After run the concurrency test, we should run that command too. If the totalAmount is same, This code works because the total balances is integrity.
*/