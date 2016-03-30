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

var countWithdraw = 0
var maxUser = 100
var maxThread = 10

//Array of channels input and output
var in []chan string
var out []chan Result

type Result struct{
  Account string
  Result string
}

func withdraw(w http.ResponseWriter, r *http.Request) {
  
  
  var wg sync.WaitGroup
  wg.Add(1)

  // random user from 1 to maxUser
  number := Random( 1, maxUser )

  //Allocate to appropriate channel number based on number by get the last number in the random number.
  channelNumber :=  number % maxThread
  /*if channelNumber == 0{
    channelNumber = 10
  }*/
  account := "user" + strconv.Itoa( Random( 1, maxUser ) )
  
  go func () {

    
    in[ channelNumber ] <- account
    for {
      select {
      case result := <- out[ channelNumber ]:

        if result.Account == account{
          /*fmt.Printf("Result %s\n", result.Result)
          fmt.Printf("Number is %d \n", channelNumber )*/
          io.WriteString(w, result.Result)
          wg.Done()
          //should return, otherwise it's still pop out value from out channel
          return          
        }else{
          fmt.Printf("Dismatch: %s and %s\n", result.Account, account)
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

  in = make([]chan string, maxThread)
  out = make([]chan Result, maxThread)

  for i := range in {

    fmt.Printf("i %d \n", i )
    in[i] = make(chan string)
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

    go func ( subIn *chan string, index int ) {

      for {
        select{
          case account := <-*subIn:
            fmt.Printf("On worker %d \n", index + 1)
            /*count_queue += 1
            fmt.Printf("count_queue %d\n", count_queue)*/
            entry := Currency{}
            err := global_db.C("bank").Find(bson.M{"account":  account }).One(&entry)
            
            //time.Sleep(100 * time.Millisecond)

            if err != nil {
              panic(err)
            }

            //fmt.Printf("%+v\n", entry)
            //step 2: check if balance is valid to widthdraw
            if entry.Amount < 50.00 {
              //fmt.Printf("out_of_balance\n")
              out[ index ] <-  Result{ Account: account, Result: "out_of_balance"}
              //io.WriteString(w, "out_of_balance")
              
            }else{
              //step 3: subtract current balance and update back to database
              entry.Amount = entry.Amount - 50.000
              err = global_db.C("bank").UpdateId(entry.Id, entry)

              if err != nil{
                //panic("update error")
                out[ index ] <-  Result{ Account: account, Result: "update error"}
              }
              countWithdraw += 1
              fmt.Printf("countWithdraw %d\n", countWithdraw)
              out[ index ] <-  Result{ Account: account, Result: fmt.Sprintf("countWithdraw %d\n", countWithdraw)}
            }
        }
      }

    }(&in[i], i)
  }

  
  http.HandleFunc("/", withdraw)
  http.ListenAndServe(":8000", nil)
}