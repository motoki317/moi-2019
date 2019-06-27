package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/marcmak/calc/calc"
)

var (
	bearer = "Bearer " + os.Getenv("ACCESS_TOKEN")
	baseUrl = "https://apiv2.twitcasting.tv/internships/2019/games"
)

type Question struct {
	ID string `json:"id"`
	Question string `json:"question"`
}

type Answer struct {
	Answer string `json:"answer"`
}

func main() {
	body, err := get(baseUrl + "?level=3")
	if err != nil {
		panic(err)
	}
	var question Question
	if err := json.Unmarshal([]byte(body), &question); err != nil {
		panic(err)
	}

	// solve

	start := time.Now()
	answer := Answer{
		Answer: solve(question),
	}
	took := float64(time.Now().Nanosecond() - start.Nanosecond()) / float64(1000000)
	fmt.Println("Answer: " + answer.Answer + ", took " + fmt.Sprintf("%f", took) + " ms.")

	body, err = post(baseUrl + "/" + question.ID, answer)
	if err != nil {
		panic(err)
	}
}

func solve(question Question) string {
	q := question.Question

	numbers := make([]int, 0)
	var stopIndex int
	for i, v := range []rune(q) {
		if n, err := strconv.Atoi(string(v)); err == nil {
			numbers = append(numbers, n)
		}
		if v == '=' {
			stopIndex = i
			break
		}
	}

	target, err := strconv.Atoi(q[stopIndex+2:])
	if err != nil {
		panic(err)
	}

	return assume(numbers, target, 1, "")
}

func assume(numbers []int, target int, depth int, current string) string {
	if depth == len(numbers) {
		if check(numbers, target, current) {
			return current
		} else {
			return ""
		}
	}

	operators := []string{"+", "-", "*", "/"}
	for _, v := range operators {
		ret := assume(numbers, target, depth+1, current + v)
		if ret != "" {
			return ret
		}
	}
	return ""
}

func check(numbers []int, target int, current string) bool {
	var toCheck string
	for i, v := range numbers {
		toCheck += strconv.Itoa(v)
		if i != len(numbers) - 1 {
			toCheck += current[i:i+1]
		}
	}
	// fmt.Println("Checking " + toCheck)

	ans := calc.Solve(toCheck)
	/*
	if int(ans) == target {
		fmt.Println("Match!")
	}
	 */
	return ans == float64(int64(ans)) && int(ans) == target
}

func get(url string) (string, error) {
	fmt.Printf("Start GET for url %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer(body)
	html := buf.String()

	fmt.Printf("Finish GET for url %s: %v, body: %s\n", url, resp.StatusCode, html)
	return html, nil
}

func post(url string, body interface{}) (string, error) {
	fmt.Printf("Start POST for url %s\n", url)
	var req *http.Request
	var err error

	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return "", err
		}
		fmt.Println("Post body: " + string(jsonBytes))
		b := bytes.NewBuffer(jsonBytes)
		req, err = http.NewRequest("POST", url, b)
		if err != nil {
			return "", err
		}

		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest("POST", url, nil)
		if err != nil {
			return "", err
		}
	}

	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	retBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer(retBody)
	html := buf.String()

	fmt.Printf("Finish POST for url %s: %v, body: %s\n", url, resp.StatusCode, html)
	return html, nil
}