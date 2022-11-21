package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Transaction struct {
	Amount    string `json:"amount"`
	TimeStamp string `json:"timestamp"`
}
type TransactionLists struct {
	T []*Transaction
}

type Statistics struct {
	Sum   string `json:"sum"`
	Avg   string `json:"avg"`
	Max   string `json:"max"`
	Min   string `json:"min"`
	Count string `json:"count"`
}
type StatsList struct {
	S *Statistics
}

func newStatistic() *Statistics {
	return &Statistics{
		Sum:   "",
		Avg:   "",
		Max:   "",
		Min:   "",
		Count: "",
	}
}

var tl TransactionLists
var st Statistics
var stats StatsList

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/transactions", CreateTransactions).Methods("POST")
	router.HandleFunc("/transactionList", FetchTransactions).Methods("GET")
	router.HandleFunc("/statistics", GetStats).Methods("GET")
	router.HandleFunc("/transactions", DeleteTransaction).Methods("DELETE")
	// router.HandleFunc("/location", createLocation).Methods("POST")
	// router.HandleFunc("/location", updateLocation).Methods("PUT")

	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8000",
	}

	log.Fatal(srv.ListenAndServe())
}

// api for creating new transaction

func CreateTransactions(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	var t Transaction
	err = json.Unmarshal(reqBody, &t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	amt, err := strconv.ParseFloat(t.Amount, 64)
	if err != nil {
		fmt.Printf("error while parsing amount field: %v", err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	//layout := time.RFC3339
	parsedTime, err := time.Parse(time.RFC3339Nano, t.TimeStamp)
	if err != nil {
		fmt.Printf("Error while parsing timestamp field: %v", err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	tx := &Transaction{
		Amount:    fmt.Sprint(amt),
		TimeStamp: parsedTime.String(),
	}
	tl.T = append(tl.T, tx)
	now := time.Now()
	fmt.Println(" now", now)
	fmt.Println("parsed after now", parsedTime.After(now))
	if !parsedTime.After(now) {
		fmt.Printf("given time %s is in future", tx.TimeStamp)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// If transaction is older than 60 sec
	if now.Sub(parsedTime) > 60*time.Second {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	statSum, _ := strconv.ParseFloat(st.Sum, 64)
	fmt.Println("stat sum is", statSum)
	statSum += amt
	st.Sum = fmt.Sprint(statSum)

	statCount, _ := strconv.ParseFloat(st.Count, 64)
	statCount += 1
	st.Count = fmt.Sprint(statCount)

	statAvg := statSum / statCount
	st.Avg = fmt.Sprint(statAvg)

	statMax, _ := strconv.ParseFloat(st.Max, 64)
	if amt > statMax {
		st.Max = fmt.Sprint(amt)
	}

	statMin, _ := strconv.ParseFloat(st.Min, 64)
	if statMin == 0 {
		st.Min = fmt.Sprint(amt)
	}
	if amt < statMin {
		st.Min = fmt.Sprint(amt)
	}
	stats.S = &st
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&tx)
	return
}

// api for get transaction
func FetchTransactions(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get call")

	if err := json.NewEncoder(w).Encode(tl.T); err != nil {
		fmt.Printf("Error while encode transactions: %v", err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	w.Header().Add("Content-Type", "application/json")
}

//api for deleting transaction
func DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	tl.T = make([]*Transaction, 0)
	st = *newStatistic()
	stats.S = &st
	w.WriteHeader(http.StatusNoContent)

}

//get call for collecting stats
func GetStats(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(stats.S); err != nil {
		fmt.Printf("Statistic is %s %s", stats.S, err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.Header().Add("Content-Type", "application/json")
}
