package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	repoName := os.Args[1]
	repoStars, err := getStars(repoName)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Stargazers count for", repoName+":", repoStars)

	fileName := "data.csv"
	csv, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer csv.Close()

	_, err = fmt.Fprintf(csv, "%s\n%d\n", repoName, repoStars)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func getStars(repoName string) (int64, error) {
	resp, err := http.Get("https://api.github.com/repos/" + repoName)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}

	stars, ok := data["stargazers_count"].(float64)
	if !ok {
		return 0, fmt.Errorf("unable to parse stargazers_count")
	}

	return int64(stars), nil
}
