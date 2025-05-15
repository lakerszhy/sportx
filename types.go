package main

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

type matchPeriod string

const (
	matchPeriodComing     matchPeriod = "0"
	matchPeriodInProgress matchPeriod = "1"
	matchPeriodEnd        matchPeriod = "2"
)

type matchType string

const (
	matchTypeFootball   matchType = "1"
	matchTypeBasketball matchType = "2"
	matchTypeSnooker    matchType = "3"
	matchTypeOther      matchType = "4"
)

type match struct {
	MID         string      `json:"mid"`
	MatchType   matchType   `json:"matchType"`
	MatchDesc   string      `json:"matchDesc"`
	StartTime   string      `json:"startTime"`
	LeftName    string      `json:"leftName"`
	LeftGoal    string      `json:"leftGoal"`
	RightName   string      `json:"rightName"`
	RightGoal   string      `json:"rightGoal"`
	MatchPeriod matchPeriod `json:"matchPeriod"`
	Quarter     string      `json:"quarter"`
	QuarterTime string      `json:"quarterTime"`
}

func (m match) isMatch() bool {
	return m.MatchType != "4"
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
	team *team
	goal *goalStatistics
}

type team struct {
	LeftName  string `json:"leftName"`
	RightName string `json:"rightName"`
}

type goalStatistics struct {
	Head []string   `json:"head"`
	Rows [][]string `json:"rows"`
}
