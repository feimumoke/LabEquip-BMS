package timeutil

import "time"

const (
	OneDaySecond               = int64(86400)
	OneHourSecond              = int64(3600)
	OneDay                     = "24h"
	DefaultTimeFormat          = "2006-01-02 15:04:05"
	HourFormat                 = "2006-01-02 15"
	DateFormat                 = "2006-01-02"
	MonthFormat                = "2006-01"
	DateIntFormat              = "20060102"
	DateIntFormatYYMMDD        = "060102"
	DatePathFormat             = "20060102"
	DateFormatDMY              = "02/01/2006"
	DateFormatUnderLineDMY     = "02-01-2006"
	DateIntFormatMMDD          = "0102"
	DateIntFormatUnderLineMMDD = "01-02"
	DateFormatMDYTime          = "1/2/2006 15:04:05"
	EastEightZone              = 8
	OneMinuteSecond            = int64(60)
)

func TodayDateStr(timeLayout string) string {
	now := time.Now().Unix()
	return TimeStampToStr(now, timeLayout)
}

func TimeStampToStr(stamp int64, format string) string {
	return time.Unix(stamp, 0).Format(format)
}

func GetCurrentUnix() int64 {
	return time.Now().Unix()
}
