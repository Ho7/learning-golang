package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"
)

type Generator struct {
	ChunkSize             int `json:"chunk_size"`
	NumbersCount          int `json:"numbers_count"`
	ThreadsCount          int `json:"threads_count"`
	generatedNumbersCount int
	numbersChain          chan int
	Numbers               []int `json:"numbers"`
}

// Init generator
func GetGenerator(numbersCount, threadChunkSize int) Generator {
	return Generator{
		threadChunkSize,
		numbersCount,
		numbersCount / threadChunkSize,
		0,
		nil,
		make([]int, numbersCount),
	}
}

func (g Generator) String() string {
	return fmt.Sprintf(
		"Generator: chunkSize=%d, numbersCount=%d, threadsCount=%d, generatedNumbersCount=%d",
		g.ChunkSize, g.NumbersCount, g.ThreadsCount, g.generatedNumbersCount)
}

// General function for generate number
func (g *Generator) SpawnNodes() {
	if g.numbersChain != nil {
		panic("new initial generator")
	}

	// spawn new thread for calculate numbers
	log.Printf("Spawn %d new threads", g.ThreadsCount)
	g.numbersChain = make(chan int, g.ThreadsCount)
	for i := 1; i <= g.ThreadsCount; i++ {
		log.Printf("New thead #%d...", i)
		go g.spawnNode(g.numbersChain, g.ChunkSize)
	}
}

// Node
func (g *Generator) spawnNode(ch chan int, count int) {
	for i := 0; i < count; i++ {
		ch <- GetRandomInt()
	}
}

func (g *Generator) SubscribeResults() {
	if g.numbersChain == nil {
		panic("run before launch SpawnNodes")
	}

	log.Printf("Subscribe for generate numbers channel")
	ticker := time.Tick(time.Second)
	for number := range g.numbersChain {
		g.Numbers[g.generatedNumbersCount] = number
		g.generatedNumbersCount++

		percent := GetPercentageOf(g.generatedNumbersCount, g.NumbersCount)
		logMessage := fmt.Sprintf("Numbers %d of %d (%v)", g.generatedNumbersCount, g.NumbersCount, percent)

		select {
		case <-ticker:
			log.Printf(logMessage)
		default:
		}
		if g.generatedNumbersCount == g.NumbersCount {
			log.Printf(logMessage)
			break
		}
	}

}

func GetRandomInt() int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Int()
}

func GetPercentageOf(count, total int) float64 {
	return math.Floor(float64(count) / float64(total) * 100)
}

func generateHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("New generating request ...")
	generator := GetGenerator(10000, 10)
	generator.SpawnNodes()
	generator.SubscribeResults()

	jsonResult, err := json.Marshal(generator)
	if err != nil {
		panic(err)
	}

	_, _ = fmt.Fprintf(w, string(jsonResult))
	log.Println("Response ...")

}

func main() {
	http.HandleFunc("/generate", generateHandler)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
