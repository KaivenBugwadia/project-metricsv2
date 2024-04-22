package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type MyDateTime struct {
	githubv4.DateTime
}

func (dt MyDateTime) ToTime() time.Time {
	return dt.DateTime.Time
}

type Repository struct {
	Packages struct {
		Nodes []struct {
			CreatedAt MyDateTime
		}
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage bool
		}
	} `graphql:"packages(first: $pageSize, after: $after)"`
}

func main() {
	repoName := os.Args[1]
	repoStars, err := getStars(repoName)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Stargazers count for", repoName+":", repoStars)

	allPackageAges, err := getAllPackageAges("KaivenBugwadia", repoName)
	if err != nil {
		log.Fatal(err)
	}

	for i, age := range allPackageAges {
		fmt.Printf("Package %d age: %s\n", i+1, age)
	}

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

func getAllPackageAges(owner, name string) ([]time.Duration, error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "YOUR_ACCESS_TOKEN_HERE"},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)

	var allPackageAges []time.Duration
	variables := map[string]interface{}{
		"owner":    githubv4.String(owner),
		"name":     githubv4.String(name),
		"pageSize": githubv4.Int(100), // Set an initial page size
	}
	for {
		var query Repository
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			return nil, err
		}

		// Calculate ages of packages for this page
		for _, pkg := range query.Packages.Nodes {
			createdAt := pkg.CreatedAt.ToTime() // Convert MyDateTime to time.Time
			allPackageAges = append(allPackageAges, time.Since(createdAt))
		}

		// Check if there are more pages
		if !query.Packages.PageInfo.HasNextPage {
			break
		}

		// Set variables for the next page
		variables["after"] = githubv4.NewString(query.Packages.PageInfo.EndCursor)
	}

	return allPackageAges, nil
}
