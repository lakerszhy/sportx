package main

import "time"

//nolint:mnd // 配置文件
var cfg = config{
	textLiveCount:             40,
	scheduleRefreshInterval:   10 * time.Second,
	statisticsRefreshInterval: 10 * time.Second,
	textLiveRefreshInterval:   5 * time.Second,
	apiRequestTimeout:         10 * time.Second,
}

type config struct {
	textLiveCount             int           // 文本直播数量
	scheduleRefreshInterval   time.Duration // 赛程刷新间隔
	statisticsRefreshInterval time.Duration // 统计刷新间隔
	textLiveRefreshInterval   time.Duration // 文本直播刷新间隔
	apiRequestTimeout         time.Duration // API请求超时时间
}
