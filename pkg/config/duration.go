package config

import (
	"encoding/json"
	"fmt"
	"time"
)

type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		parsed, err := time.ParseDuration(str)
		if err != nil {
			return fmt.Errorf("invalid duration string %q: %w", str, err)
		}
		*d = Duration(parsed)
		return nil
	}

	var num int64
	if err := json.Unmarshal(data, &num); err != nil {
		return fmt.Errorf("duration must be a string (e.g., \"2m\") or number (nanoseconds)")
	}
	*d = Duration(num)
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}
