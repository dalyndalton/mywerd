package words

import (
	"bufio"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func ExponentialRandom() int {
	// Generate a random number between 0 and 1
	u := rand.ExpFloat64() * 2.5

	// Scale the value to the desired range (0 to 100)
	if u > 100 {
		u = 100
	}
	if u < 10 {
		u = 10
	}

	return int(u)
}

var wordLists map[int][]string
var wordWeights []int

func loadWords() {
	files, err := os.ReadDir("/usr/share/dict/scowl")
	if err != nil {
		panic(err)
	}

	wordLists = make(map[int][]string)

	for _, fileInfo := range files {
		if !fileInfo.IsDir() && strings.HasPrefix(fileInfo.Name(), "english-") {
			parts := strings.Split(fileInfo.Name(), ".")
			if len(parts) != 2 {
				continue
			}

			number, err := strconv.Atoi(parts[1])
			if err != nil {
				continue
			}

			file, err := os.Open("/usr/share/dict/scowl/" + fileInfo.Name())
			if err != nil {
				panic(err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				wordLists[number] = append(wordLists[number], scanner.Text())
			}

			if err := scanner.Err(); err != nil {
				panic(err)
			}
		}
	}

	totalWords := 0
	for _, words := range wordLists {
		totalWords += len(words)
	}

	wordWeights = make([]int, 0, len(wordLists))
	for k := range wordLists {
		wordWeights = append(wordWeights, k)
	}
	sort.Ints(wordWeights)
	logrus.Infof("Total words loaded: %d", totalWords)
}

type Word struct {
	Word   string `json:"word"`
	Rarity int    `json:"rarity"`
}

func RandomWordsHandler(w http.ResponseWriter, r *http.Request) {

	selectedWords := make([]Word, 5)
	for i := 0; i < 5; i++ {
		word, rarity := RandomWord()
		selectedWords[i] = Word{Word: word, Rarity: rarity}
	}

	jsonResponse, err := json.Marshal(selectedWords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func RandomWord() (string, int) {
	randValue := ExponentialRandom()
	rarity := 0

	for _, weight := range wordWeights {
		if weight > randValue {
			break
		}
		rarity = weight
	}

	wordListLength := len(wordLists[rarity])
	return wordLists[rarity][rand.Intn(wordListLength)], rarity
}

func init() {
	loadWords()
}
