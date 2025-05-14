package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
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
			categories = append(categories, c)
		}
	}
	return categories, nil
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

func fetchTextLives(matchID string) ([]textLive, error) {
	// https://matchweb.sports.qq.com/kbs/matchDetail?mid=100000:10042400225&from=sportsh5
	// 是否有文字直播
	indexs, err := fetchTextLiveIndexes(matchID)
	if err != nil {
		return nil, err
	}

	if len(indexs) == 0 {
		return nil, nil
	}

	if len(indexs) > 40 {
		indexs = indexs[:40]
	}

	ret, err := fetchIndexTexts(matchID, indexs)
	if err != nil {
		return nil, err
	}

	var textLives []textLive
	for _, index := range indexs {
		textLives = append(textLives, ret[index])
	}

	slices.SortStableFunc(textLives, func(a, b textLive) int {
		indexA := strings.Split(a.IndexValue, "_")
		indexB := strings.Split(b.IndexValue, "_")
		if len(indexA) != 2 || len(indexB) != 2 {
			return 0
		}
		retA, err := strconv.Atoi(indexA[0])
		if err != nil {
			return 0
		}
		retB, err := strconv.Atoi(indexB[0])
		if err != nil {
			return 0
		}
		return retA - retB
	})

	return textLives, nil
}

func fetchTextLiveIndexes(matchID string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet,
		"https://app.sports.qq.com/textLive/index",
		nil,
	)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("mid", matchID)

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")
	req.URL.RawQuery = params.Encode()

	hresp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer hresp.Body.Close()

	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Tabs []struct {
				TabName string   `json:"tabName"`
				Index   []string `json:"index"`
			} `json:"tabs"`
		} `json:"data"`
	}
	if err := json.NewDecoder(hresp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("fetch text live index failed, code: %d, msg: %s",
			resp.Code, resp.Msg)
	}

	if len(resp.Data.Tabs) == 0 {
		return nil, nil
	}

	return resp.Data.Tabs[0].Index, nil
}

func fetchIndexTexts(matchID string, indexes []string) (map[string]textLive, error) {
	ids := strings.Split(matchID, ":")
	if len(ids) != 2 {
		return nil, fmt.Errorf("invalid match id: %s", matchID)
	}

	req, err := http.NewRequest(http.MethodGet,
		"https://matchweb.sports.qq.com/textLive/detail",
		nil,
	)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("competitionId", ids[0])
	params.Add("matchId", ids[1])
	params.Add("ids", strings.Join(indexes, ","))

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")
	req.URL.RawQuery = params.Encode()

	hresp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer hresp.Body.Close()

	var resp []json.RawMessage
	if err := json.NewDecoder(hresp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if len(resp) != 3 {
		return nil, fmt.Errorf("fetch text live failed, resp: %v", resp)
	}

	var ret map[string]textLive
	if err := json.Unmarshal(resp[1], &ret); err != nil {
		return nil, err
	}

	return ret, nil
}
