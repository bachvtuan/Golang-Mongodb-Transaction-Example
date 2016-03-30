package main

import (
  //"strconv"
  "sync"
  "io"
  "net/http"
  "fmt"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  //"time"
)


var global_db *mgo.Database

var mu = &sync.Mutex{}

type Currency struct {
  Id   bson.ObjectId `json:"id" bson:"_id,omitempty"`
  Amount float64 `bson:"amount"`
  Account   string `bson:"account"`
  Code   string `bson:"code"`
}

var count_withdraw = 0

func withdraw(w http.ResponseWriter, r *http.Request) {
  entry := Currency{}
  //step 1: get current amount

  //Solution here, Lock other thread access this section of code until it's unlocked
  mu.Lock()
  defer mu.Unlock()
  err := global_db.C("bank").Find(bson.M{"account":  "tuanbach"}).One(&entry)

  if err != nil {
    panic(err)
  }

  fmt.Printf("%+v\n", entry)
  //step 2: check if balance is valid to widthdraw
  if entry.Amount < 50.00 {
    fmt.Printf("out_of_balance %d\n", count_withdraw)
    io.WriteString(w, "out_of_balance")
    return
  }

  //step 3: subtract current balance and update back to database
  entry.Amount = entry.Amount - 50.000
  err = global_db.C("bank").UpdateId(entry.Id, entry)

  if err != nil{
    panic("update error")
  }
  count_withdraw += 1
  fmt.Printf("count_withdraw %d\n", count_withdraw)

  io.WriteString(w, "ok")

}

func main() {
  session, _ := mgo.Dial("localhost:27017")
  fmt.Printf("Session is %p\n", session)
  global_db = session.DB( "db_log" )

  //make sure it is empty first
  global_db.C("bank").DropCollection()

  //Init amount is 1000 USD
  user := Currency{Account : "tuanbach", Amount: 1000.00, Code:"USD"}
  err := global_db.C("bank").Insert(&user)

  if err != nil{
    panic("insert error")
  }

  http.HandleFunc("/", withdraw)
  http.ListenAndServe(":8000", nil)
}