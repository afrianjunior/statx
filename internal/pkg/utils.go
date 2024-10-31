package pkg

import "time"

func ParseTimeRange(start string, end string, duration string) (TimeRange, error) {
	now := time.Now()
	defaultRange := TimeRange{
		Start: now.Add(-1 * time.Hour),
		End:   now,
	}

	if duration != "" {
		d, err := time.ParseDuration(duration)
		if err != nil {
			return defaultRange, err
		}
		return TimeRange{
			Start: now.Add(-d),
			End:   now,
		}, nil
	}

	if start != "" && end != "" {
		startTime, err := time.Parse(time.RFC3339, start)
		if err != nil {
			return defaultRange, err
		}
		endTime, err := time.Parse(time.RFC3339, end)
		if err != nil {
			return defaultRange, err
		}
		return TimeRange{Start: startTime, End: endTime}, nil
	}

	return defaultRange, nil
}
