package formats

import (
	"encoding/json"
	"fmt"
	"time"
)

var dateTimeFormats = []string{
	"2006-01-02 15:04:05.999999",
	"2006-01-02T15:04:05.999999",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
}

// useMicroSeconds tells whether the marshalling
// dumps JSON with nanoseconds part or not.
var useMicroSeconds = false

// Time works like time but has a custom formatting.
type Time time.Time

// UseMicroSeconds sets whether this application (or set of)
// should use microseconds while dumping date/times or not.
func UseMicroSeconds(use bool) {
	useMicroSeconds = use
}

// MarshalJSON does a JSON marshalling.
func (ct *Time) MarshalJSON() ([]byte, error) {
	var dateTimeFormatIndex = 0
	if useMicroSeconds {
		dateTimeFormatIndex = 1
	}
	return json.Marshal(time.Time(*ct).Format(dateTimeFormats[dateTimeFormatIndex]))
}

// UnmarshalJSON does a JSON de-marshalling through all
// the available date/time formats until one matches.
func (ct *Time) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	for _, dateTimeFormat := range dateTimeFormats {
		t, err := time.Parse(dateTimeFormat, str)
		if err == nil {
			*ct = Time(t)
			return nil
		}
	}
	return fmt.Errorf("time data '%s' does not match any of the available formats", str)
}
