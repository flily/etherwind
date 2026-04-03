package ping

import (
	"errors"
	"math"
	"os"
)

func rootError(err error) error {
	for {
		next := errors.Unwrap(err)
		if next == nil {
			return err
		}
		err = next
	}
}

func isPermissionDenied(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, os.ErrPermission) {
		return true
	}

	return false
}

type TimeRecord struct {
	Records []float64
}

func NewTimeRecords(size int) *TimeRecord {
	r := &TimeRecord{
		Records: make([]float64, 0, size),
	}

	return r
}

func (r *TimeRecord) Add(value float64) {
	r.Records = append(r.Records, value)
}

func (r *TimeRecord) Min() float64 {
	if len(r.Records) == 0 {
		return 0
	}

	min := r.Records[0]
	for _, v := range r.Records {
		if v < min {
			min = v
		}
	}

	return min
}

func (r *TimeRecord) Max() float64 {
	if len(r.Records) == 0 {
		return 0
	}

	max := r.Records[0]
	for _, v := range r.Records {
		if v > max {
			max = v
		}
	}

	return max
}

func (r *TimeRecord) Length() int {
	return len(r.Records)
}

func (r *TimeRecord) Sum() float64 {
	sum := 0.0
	for _, v := range r.Records {
		sum += v
	}

	return sum
}

func (r *TimeRecord) Average() float64 {
	if len(r.Records) == 0 {
		return 0
	}

	return r.Sum() / float64(r.Length())
}

func (r *TimeRecord) StandardDeviation() float64 {
	if len(r.Records) == 0 {
		return 0
	}

	avg := r.Average()
	varianceSum := 0.0
	for _, v := range r.Records {
		diff := v - avg
		varianceSum += diff * diff
	}

	variance := varianceSum / float64(r.Length())
	return math.Sqrt(variance)
}
