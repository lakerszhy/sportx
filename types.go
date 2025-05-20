package main

import (
	"github.com/charmbracelet/x/ansi"
)

var hotCategory = category{
	ID:   "hot",
	Name: "热门",
}

type category struct {
	ID   string `json:"columnId"`
	Name string `json:"name"`
}

func (c category) FilterValue() string {
	return c.Name
}

func (c category) equal(v category) bool {
	return c.ID == v.ID
}

type period string

const (
	periodComing     period = "0"
	periodInProgress period = "1"
	periodEnd        period = "2"
)

type matchType string

// const (
// 	matchTypeFootball   matchType = "1"
// 	matchTypeBasketball matchType = "2"
// 	matchTypeSnooker    matchType = "3"
// 	matchTypeOther      matchType = "4"
// )

type match struct {
	MID         string    `json:"mid"`
	MatchType   matchType `json:"matchType"`
	MatchDesc   string    `json:"matchDesc"`
	StartTime   string    `json:"startTime"`
	LeftName    string    `json:"leftName"`
	LeftGoal    string    `json:"leftGoal"`
	RightName   string    `json:"rightName"`
	RightGoal   string    `json:"rightGoal"`
	MatchPeriod period    `json:"matchPeriod"`
	Quarter     string    `json:"quarter"`
	QuarterTime string    `json:"quarterTime"`
}

func (m match) isMatch() bool {
	return m.MatchDesc != "NBA经典赛" &&
		m.MatchDesc != "篮球直播节目" &&
		m.MatchDesc != "发布会"
}

func (m match) FilterValue() string {
	return ""
}

type textLive struct {
	Content    string `json:"content"`
	LeftGoal   string `json:"leftGoal"`
	RightGoal  string `json:"rightGoal"`
	IndexValue string `json:"indexValue"`
	Plus       string `json:"plus"`
	Quarter    struct {
		Quarter string `json:"quarter"`
		Time    string `json:"time"`
	} `json:"kbsInfo"`
}

func (t textLive) FilterValue() string {
	return t.Content
}

type statistics struct {
	team             *team
	goal             *goalStatistics
	teamStatistics   []teamStatistics
	playerStatistics [][]playerStatistics
	livePeriod       period
}

type team struct {
	LeftName  string `json:"leftName"`
	RightName string `json:"rightName"`
}

func (t team) width() int {
	left := ansi.StringWidth(t.LeftName)
	right := ansi.StringWidth(t.RightName)
	if left > right {
		return left
	}
	return right
}

type goalStatistics struct {
	Head []string   `json:"head"`
	Rows [][]string `json:"rows"`
}

type teamStatistics struct {
	LeftVal  string `json:"leftVal"`
	RightVal string `json:"rightVal"`
	Text     string `json:"text"`
}

type playerStatistics struct {
	Head []string `json:"head"`
	Row  []string `json:"row"`
}
