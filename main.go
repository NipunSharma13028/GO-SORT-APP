package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"
)

type RequestPayload struct {
	ToSort [][]int `json:"to_sort"`
}

type ResponsePayload struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNS       int64   `json:"time_ns"`
}

func main() {
	http.HandleFunc("/process-single", handleSequentialProcessing)
	http.HandleFunc("/process-concurrent", handleConcurrentProcessing)

	http.ListenAndServe(":8000", nil)
}

func handleSequentialProcessing(w http.ResponseWriter, r *http.Request) {
	processAndRespond(w, r, false)
}

func handleConcurrentProcessing(w http.ResponseWriter, r *http.Request) {
	processAndRespond(w, r, true)
}

func processAndRespond(w http.ResponseWriter, r *http.Request, concurrent bool) {
	var reqPayload RequestPayload
	err := json.NewDecoder(r.Body).Decode(&reqPayload)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	startTime := time.Now()

	var sortedArrays [][]int
	if concurrent {
		sortedArrays = sortConcurrently(reqPayload.ToSort)
	} else {
		sortedArrays = sortSequentially(reqPayload.ToSort)
	}

	elapsedTime := time.Since(startTime)

	responsePayload := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNS:       elapsedTime.Nanoseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responsePayload)
}

func sortSequentially(arrays [][]int) [][]int {
	var result [][]int

	for _, arr := range arrays {
		sortedArr := make([]int, len(arr))
		copy(sortedArr, arr)
		sort.Ints(sortedArr)
		result = append(result, sortedArr)
	}

	return result
}

func sortConcurrently(arrays [][]int) [][]int {
	var result [][]int
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, arr := range arrays {
		wg.Add(1)
		go func(arr []int) {
			defer wg.Done()
			sortedArr := make([]int, len(arr))
			copy(sortedArr, arr)
			sort.Ints(sortedArr)

			mu.Lock()
			result = append(result, sortedArr)
			mu.Unlock()
		}(arr)
	}

	wg.Wait()
	return result
}
