package formats

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math"
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

// DateTimeInvalid stands for an invalid date/time.
const DateTimeInvalid DateTime = math.MinInt64

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

// parseTime parses the time string into a value to
// assign to the current time.
func (ct *DateTime) parseTime(str string) error {
	for _, dateTimeFormat := range dateTimeFormats {
		t, err := time.Parse(dateTimeFormat, str)
		if err == nil {
			*ct = DateTime(primitive.NewDateTimeFromTime(t))
			return nil
		}
	}
	return fmt.Errorf("time data '%s' does not match any of the available formats", str)
}

// UnmarshalJSON does a JSON de-marshalling through all
// the available date/time formats until one matches.
func (ct *DateTime) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	return ct.parseTime(str)
}

// MarshalBSONValue forwards its logic to the given base
// implementation for primitive.DateTime.
func (ct *DateTime) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(primitive.DateTime(*ct))
}

// UnmarshalBSONValue is implemented through access to the
// string value (if string) or the BSON DateTime value (if
// date-time in database). This serves well for the purpose
// both for conversions and for simulated updates.
func (ct *DateTime) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t == bson.TypeString {
		var str string
		if err := bson.UnmarshalValue(t, data, &str); err != nil {
			*ct = DateTimeInvalid
			return err
		}
		return ct.parseTime(str)
	} else {
		var v primitive.DateTime
		if err := bson.UnmarshalValue(t, data, &v); err != nil {
			return err
		} else {
			*ct = DateTime(v)
			return nil
		}
	}
}
