package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"time"
)

func fetchCategories() ([]category, error) {
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Title      string     `json:"title"`
			ShowLimit  string     `json:"showLimit"`
			Categories []category `json:"columns"`
		} `json:"data"`
	}

	hresp, err := http.Get("https://matchweb.sports.qq.com/matchUnion/cateColumns")
	if err != nil {
		return nil, err
	}
	defer hresp.Body.Close()

	if hresp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch categories failed: %d", hresp.StatusCode)
	}

	if err := json.NewDecoder(hresp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("fetch categories failed, code: %d, msg: %s",
			resp.Code, resp.Msg)
	}

	categories := []category{hotCategory}
	for _, v := range resp.Data {
		for _, c := range v.Categories {
			c.fetchSchedule = func() ([]match, error) {
				return fetchSchedule(c.ID)
			}
			categories = append(categories, c)
		}
	}
	return categories, nil
}

func fetchHotMatches() ([]match, error) {
	resp, err := http.Get("https://matchweb.sports.qq.com/html/hotMatchList")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch hot match failed, status code: %d", resp.StatusCode)
	}

	var respData []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, err
	}

	if len(respData) < 2 {
		return nil, nil
	}

	var code int
	if err := json.Unmarshal(respData[0], &code); err != nil {
		return nil, err
	}
	if code != 0 {
		return nil, fmt.Errorf("fetch hot match failed, code: %d", code)
	}

	var data map[string][]match
	if err := json.Unmarshal(respData[1], &data); err != nil {
		return nil, err
	}

	return sortMatches(data)
}

func fetchSchedule(categoyID string) ([]match, error) {
	var resp struct {
		Code int                `json:"code"`
		Msg  string             `json:"msg"`
		Data map[string][]match `json:"data"`
	}

	req, err := http.NewRequest(http.MethodGet,
		"https://matchweb.sports.qq.com/matchUnion/list",
		nil,
	)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	end := start.AddDate(0, 0, 7)
	params := url.Values{}
	params.Add("columnId", categoyID)
	params.Add("startTime", start.Format("2006-01-02"))
	params.Add("endTime", end.Format("2006-01-02"))

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")
	req.URL.RawQuery = params.Encode()

	hresp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer hresp.Body.Close()

	if hresp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch schedule failed: %d", hresp.StatusCode)
	}

	if err := json.NewDecoder(hresp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("fetch schedule failed, code: %d, msg: %s",
			resp.Code, resp.Msg)
	}

	return sortMatches(resp.Data)
}

func sortMatches(data map[string][]match) ([]match, error) {
	type dailyMatchers struct {
		date    time.Time
		matches []match
	}

	var daily []dailyMatchers
	for k, v := range data {
		date, err := time.Parse("2006-01-02", k)
		if err != nil {
			continue
		}
		daily = append(daily, dailyMatchers{
			date:    date,
			matches: v,
		})
	}

	slices.SortStableFunc(daily, func(a, b dailyMatchers) int {
		return a.date.Compare(b.date)
	})

	var matches []match
	for _, v := range daily {
		for _, m := range v.matches {
			if m.isMatch() {
				matches = append(matches, m)
			}
		}
	}

	return matches, nil
}
