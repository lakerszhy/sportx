package main

const (
	statusInitial status = iota
	statusLoading
	statusSuccess
	statusFailed
)

type status int

func (s status) isInitial() bool {
	return s == statusInitial
}

func (s status) isLoading() bool {
	return s == statusLoading
}

func (s status) isSuccess() bool {
	return s == statusSuccess
}

func (s status) isFailed() bool {
	return s == statusFailed
}

type categoriesMsg struct {
	categories []category
	err        error
	status
}

func newCategoriesLoadingMsg() categoriesMsg {
	return categoriesMsg{
		status: statusLoading,
	}
}

func newCategoriesLoadedMsg(categories []category) categoriesMsg {
	return categoriesMsg{
		categories: categories,
		status:     statusSuccess,
	}
}

func newCategoriesFailedMsg(err error) categoriesMsg {
	return categoriesMsg{
		err:    err,
		status: statusFailed,
	}
}

type categorySelectionMsg category

type scheduleMsg struct {
	category category
	matches  []match
	err      error
	status
}

func newScheduleInitialMsg() scheduleMsg {
	return scheduleMsg{
		status: statusInitial,
	}
}

func newScheduleLoadingMsg(category category) scheduleMsg {
	return scheduleMsg{
		category: category,
		status:   statusLoading,
	}
}

func newScheduleLoadedMsg(category category, matches []match) scheduleMsg {
	return scheduleMsg{
		category: category,
		matches:  matches,
		status:   statusSuccess,
	}
}

func newScheduleFailedMsg(category category, err error) scheduleMsg {
	return scheduleMsg{
		category: category,
		err:      err,
		status:   statusFailed,
	}
}

type matchSelectionMsg string

type textLivesMsg struct {
	matchID   string
	textLives []textLive
	hasData   bool
	err       error
	status
}

func newTextLivesInitialMsg() textLivesMsg {
	return textLivesMsg{
		status: statusInitial,
	}
}

func newTextLivesLoadingMsg(matchID string) textLivesMsg {
	return textLivesMsg{
		matchID: matchID,
		status:  statusLoading,
	}
}

func newTextLivesLoadedMsg(matchID string, textLives []textLive) textLivesMsg {
	return textLivesMsg{
		matchID:   matchID,
		textLives: textLives,
		hasData:   true,
		status:    statusSuccess,
	}
}

func newTextLivesNoDataMsg(matchID string) textLivesMsg {
	return textLivesMsg{
		matchID: matchID,
		hasData: false,
		status:  statusSuccess,
	}
}

func newTextLivesFailedMsg(matchID string, err error) textLivesMsg {
	return textLivesMsg{
		matchID: matchID,
		hasData: true,
		err:     err,
		status:  statusFailed,
	}
}

type statisticsMsg struct {
	matchID    string
	statistics *statistics
	err        error
	status
}

func newStatisticsInitialMsg() statisticsMsg {
	return statisticsMsg{
		status: statusInitial,
	}
}

func newStatisticsLoadingMsg(matchID string) statisticsMsg {
	return statisticsMsg{
		matchID: matchID,
		status:  statusLoading,
	}
}

func newStatisticsLoadedMsg(matchID string, s *statistics) statisticsMsg {
	return statisticsMsg{
		matchID:    matchID,
		statistics: s,
		status:     statusSuccess,
	}
}

func newStatisticsFailedMsg(matchID string, err error) statisticsMsg {
	return statisticsMsg{
		matchID: matchID,
		err:     err,
		status:  statusFailed,
	}
}
