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
	Categories []category
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
		Categories: categories,
		status:     statusSuccess,
	}
}

func newCategoriesFailedMsg(err error) categoriesMsg {
	return categoriesMsg{
		err:    err,
		status: statusFailed,
	}
}
