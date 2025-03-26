package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

func ParseDateToISO(date string) (time.Time, error) {
	var layout = os.Getenv("DATE_LAYOUT")
	d, err := time.Parse(layout, date)
	log.Default().Println(layout, d)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed parsing date: %v", err)
	}
	return d, nil
}
