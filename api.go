package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	var categories []category
	for _, v := range resp.Data {
		for _, c := range v.Categories {
			categories = append(categories, c)
		}
	}
	return categories, nil
}
