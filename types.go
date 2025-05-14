package main

type category struct {
	ID   string `json:"columnId"`
	Name string `json:"name"`
}

func (c category) FilterValue() string {
	return c.Name
}
