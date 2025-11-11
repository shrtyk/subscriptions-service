package http

import "time"

type DateLayout string

func (dl DateLayout) parse(dateStr string) (time.Time, error) {
	return time.Parse(string(dateLayout), dateStr)
}

func (dl DateLayout) format(date time.Time) string {
	return date.Format(string(dl))
}

const dateLayout DateLayout = "01-2006"
