package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

const HistoryFilename = "history.txt"
const Delimiter = "######"

type TranslatedMessage struct {
	GopherWord string `json:"gopher-word"`
}

func main() {
	portFlag := flag.String("port", "8080", "optional port")
	flag.Parse()

	http.HandleFunc("/word/", wordHandler)
	http.HandleFunc("/history/", historyHandler)
	http.HandleFunc("/sentence/", sentenceHandler)
	log.Fatal(http.ListenAndServe(":"+*portFlag, nil))
}

func wordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprint(w, "POST method expected")
		log.Fatalln("POST method expected")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}

	m := struct {
		EnglishWord string `json:"english-word"`
	}{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Fatalln(err)
	}

	translatedWord := translate(m.EnglishWord)
	//log history record
	writeHistory(m.EnglishWord + Delimiter + translatedWord)

	tm := TranslatedMessage{translatedWord}
	jsonResult, err := json.Marshal(tm)
	if err != nil {
		log.Fatalln(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)
}

func sentenceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprint(w, "POST method expected")
		log.Fatalln("POST method expected")
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}

	m := struct {
		EnglishSentence string `json:"english-sentence"`
	}{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Fatalln(err)
	}

	m.EnglishSentence = strings.TrimSpace(m.EnglishSentence)

	//end punctuation
	endSign := m.EnglishSentence[len(m.EnglishSentence)-1:]

	sentenceWords := strings.Split(m.EnglishSentence[:len(m.EnglishSentence)-1], " ")

	var translatedWords []string
	for i, word := range sentenceWords {
		if i == 0 {
			//capitalize first word
			translatedWords = append(translatedWords, strings.Title(translate(word)))
		} else {
			translatedWords = append(translatedWords, translate(word))
		}
	}

	//join the words to form a sentence
	translatedSentence := strings.Join(translatedWords, " ")
	translatedSentence = translatedSentence + endSign

	//log history record
	writeHistory(m.EnglishSentence + Delimiter + translatedSentence)

	tm := struct {
		GopherSentence string `json:"gopher-sentence"`
	}{translatedSentence}

	jsonResult, err := json.Marshal(tm)
	if err != nil {
		log.Fatalln(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	historyRecords := readHistory()

	jsonResult, err := json.Marshal(historyRecords)
	if err != nil {
		log.Fatalln(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)
}

func writeHistory(record string) {
	filename := HistoryFilename
	info, err := os.Stat(filename)

	if !os.IsNotExist(err) && !info.IsDir() {
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln(err)
		}
		defer file.Close()
		//append to file
		if _, err := file.WriteString("\n" + record); err != nil {
			log.Fatalln(err)
		}
	} else {
		//create file
		err := ioutil.WriteFile(filename, []byte(record), 0644)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func readHistory() map[string][]map[string]string {
	var data = make(map[string][]map[string]string)

	filename := HistoryFilename

	info, err := os.Stat(filename)
	if !os.IsNotExist(err) && !info.IsDir() {
		body, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalln(err)
		}

		//split lines by new line
		lines := strings.Split(string(body), "\n")
		//sort lines alphabetically
		sort.Slice(lines, func(i, j int) bool { return strings.ToLower(lines[i]) < strings.ToLower(lines[j]) })

		data["history"] = make([]map[string]string, len(lines))

		for i, line := range lines {
			records := strings.Split(line, Delimiter)
			data["history"][i] = make(map[string]string)
			data["history"][i][records[0]] = records[1]
		}
	}
	return data
}

func findVowel(word string) (int, bool) {
	vowels := []string{"a", "e", "o", "i", "u", "y"}

	for i, letter := range word {
		for _, vowel := range vowels {
			if string(letter) == vowel {
				return i, true
			}
		}
	}
	return -1, false
}

func translate(word string) (translation string) {
	word = strings.ToLower(word)
	vowelPos, result := findVowel(word)

	if !result {
		return word
	}

	switch vowelPos {
	case 0:
		//"apple” -> “gapple”
		translation = "g" + word
	default:
		switch {
		//"xray -> gexray
		case word[0:2] == "xr":
			translation = "ge" + word
		//"square" -> "aresquogo
		case vowelPos == 2 && word[1:3] == "qu":
			translation = word[3:] + word[:3] + "ogo"
		//"chair" -> "airchogo”
		default:
			translation = word[vowelPos:] + word[:vowelPos] + "ogo"
		}
	}
	return translation
}
