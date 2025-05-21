package main

import "time"

var cfg = config{
	textLiveCount:             40,
	scheduleRefreshInterval:   10 * time.Second,
	statisticsRefreshInterval: 10 * time.Second,
	textLiveRefreshInterval:   5 * time.Second,
}

type config struct {
	textLiveCount             int           // 文本直播数量
	scheduleRefreshInterval   time.Duration // 赛程刷新间隔
	statisticsRefreshInterval time.Duration // 统计刷新间隔
	textLiveRefreshInterval   time.Duration // 文本直播刷新间隔
}
