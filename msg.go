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
