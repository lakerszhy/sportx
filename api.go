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

	err := request(
		"https://matchweb.sports.qq.com/matchUnion/cateColumns",
		nil,
		&resp,
	)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("fetch categories failed, code: %d, msg: %s",
			resp.Code, resp.Msg)
	}

	categories := []category{hotCategory}
	for _, v := range resp.Data {
		categories = append(categories, v.Categories...)
	}
	return categories, nil
}

func fetchSchedule(categoyID string) ([]match, error) {
	var resp struct {
		Code int                `json:"code"`
		Msg  string             `json:"msg"`
		Data map[string][]match `json:"data"`
	}

	start := time.Now()
	p := map[string]string{
		"columnId":  categoyID,
		"startTime": start.Format("2006-01-02"),
		"endTime":   start.AddDate(0, 0, 5).Format("2006-01-02"),
	}
	err := request(
		"https://matchweb.sports.qq.com/matchUnion/list",
		p,
		&resp,
	)
	if err != nil {
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
	indexs, err := fetchTextLiveIndexes(matchID)
	if err != nil {
		return nil, err
	}

	if len(indexs) == 0 {
		return nil, nil
	}

	if len(indexs) > cfg.textLiveCount {
		indexs = indexs[:cfg.textLiveCount]
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

func fetchMatchHasTextLives(matchID string) (bool, error) {
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			IsHasTextLive bool `json:"isHasTextLive"`
		} `json:"data"`
	}

	err := request(
		"https://matchweb.sports.qq.com/kbs/matchDetail",
		map[string]string{"mid": matchID},
		&resp,
	)
	if err != nil {
		return false, err
	}

	if resp.Code != 0 {
		return false, fmt.Errorf("fetch match has text live sfailed, code: %d, msg: %s",
			resp.Code, resp.Msg)
	}

	return resp.Data.IsHasTextLive, nil
}

func fetchTextLiveIndexes(matchID string) ([]string, error) {
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

	err := request(
		"https://app.sports.qq.com/textLive/index",
		map[string]string{"mid": matchID},
		&resp,
	)
	if err != nil {
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

	var resp []json.RawMessage

	p := map[string]string{
		"competitionId": ids[0],
		"matchId":       ids[1],
		"ids":           strings.Join(indexes, ","),
	}
	err := request(
		"https://matchweb.sports.qq.com/textLive/detail",
		p,
		&resp,
	)
	if err != nil {
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

func fetchStatistics(matchID string) (*statistics, error) {
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			TeamInfo   team   `json:"teamInfo"`
			LivePeriod period `json:"livePeriod"`
			Stats      []struct {
				Type        string           `json:"type"`
				Goals       []goalStatistics `json:"goals"`
				TeamStats   []teamStatistics `json:"teamStats"`
				PlayerStats json.RawMessage  `json:"playerStats"`
			} `json:"stats"`
		} `json:"data"`
	}

	err := request(
		"https://app.sports.qq.com/match/statDetail",
		map[string]string{"mid": matchID},
		&resp,
	)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("fetch statistics failed, code: %d, msg: %s",
			resp.Code, resp.Msg)
	}

	var g *goalStatistics
	var t []teamStatistics
	var p []playerStatistics
	for _, v := range resp.Data.Stats {
		if v.Type == "12" && len(v.Goals) > 0 {
			g = &v.Goals[0]
		} else if v.Type == "14" || v.Type == "102" {
			// 14：篮球 102：足球
			t = v.TeamStats
		} else if v.Type == "15" {
			err := json.Unmarshal(v.PlayerStats, &p)
			if err != nil {
				return nil, err
			}
		}
	}

	return &statistics{
		team:             &resp.Data.TeamInfo,
		goal:             g,
		teamStatistics:   t,
		livePeriod:       resp.Data.LivePeriod,
		playerStatistics: splitPlayerStatistics(p),
	}, nil
}

func splitPlayerStatistics(s []playerStatistics) [][]playerStatistics {
	var teams [][]playerStatistics
	var players []playerStatistics

	for _, v := range s {
		if len(v.Head) == 0 && len(v.Row) == 0 {
			continue
		}

		if len(v.Head) > 0 && len(players) > 0 {
			teams = append(teams, players)
			players = []playerStatistics{}
		}
		players = append(players, v)
	}
	teams = append(teams, players)

	return teams
}

func request(u string, p map[string]string, ret any) error {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	q := url.Values{}
	for k, v := range p {
		q.Add(k, v)
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")
	req.URL.RawQuery = q.Encode()

	hresp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer hresp.Body.Close() //nolint: errcheck

	if hresp.StatusCode != http.StatusOK {
		return fmt.Errorf("request %s failed: %d", u, hresp.StatusCode)
	}

	return json.NewDecoder(hresp.Body).Decode(&ret)
}
