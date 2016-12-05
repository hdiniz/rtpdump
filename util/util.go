package util

import (
  "fmt"
  "time"
)

var timeFormat = "%02d-%02d-%d %02d:%02d:%02d"

func TimeToStr(t time.Time) string {
  return fmt.Sprintf(
      timeFormat,
      t.UTC().Day(),
      t.UTC().Month(),
      t.UTC().Year(),
      t.UTC().Hour(),
      t.UTC().Minute(),
      t.UTC().Second())
}
