package formats

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// DateTime works like time but has a custom formatting.
type DateTime primitive.DateTime

// UseMicroSeconds sets whether this application (or set of)
// should use microseconds while dumping date/times or not.
func UseMicroSeconds(use bool) {
	useMicroSeconds = use
}

// MarshalJSON does a JSON marshalling.
func (ct DateTime) MarshalJSON() ([]byte, error) {
	var dateTimeFormatIndex = 0
	if useMicroSeconds {
		dateTimeFormatIndex = 1
	}
	return json.Marshal(primitive.DateTime(ct).Time().Format(dateTimeFormats[dateTimeFormatIndex]))
}

// UnmarshalJSON does a JSON de-marshalling through all
// the available date/time formats until one matches.
func (ct *DateTime) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	for _, dateTimeFormat := range dateTimeFormats {
		t, err := time.Parse(dateTimeFormat, str)
		if err == nil {
			*ct = DateTime(primitive.NewDateTimeFromTime(t))
			return nil
		}
	}
	return fmt.Errorf("time data '%s' does not match any of the available formats", str)
}
