package ping

import (
	"errors"
	"math"
	"os"
	"time"
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

type PingRecord struct {
	Start   time.Time
	Finish  time.Time
	Count   int
	Success int
	Records []float64
}

func NewPingRecord(size int) *PingRecord {
	t := time.Now()
	r := &PingRecord{
		Start:   t,
		Finish:  t,
		Count:   0,
		Success: 0,
		Records: make([]float64, 0, size),
	}

	return r
}

func (r *PingRecord) Add(value float64) {
	r.Records = append(r.Records, value)
	r.Count++
	r.Finish = time.Now()
}

func (r *PingRecord) AddFailure() {
	r.Count++
	r.Finish = time.Now()
}

func (r *PingRecord) Min() float64 {
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

func (r *PingRecord) Max() float64 {
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

func (r *PingRecord) Length() int {
	return len(r.Records)
}

func (r *PingRecord) Sum() float64 {
	sum := 0.0
	for _, v := range r.Records {
		sum += v
	}

	return sum
}

func (r *PingRecord) Average() float64 {
	if len(r.Records) == 0 {
		return 0
	}

	return r.Sum() / float64(r.Length())
}

func (r *PingRecord) StandardDeviation() float64 {
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

func (r *PingRecord) PacketsSuccess() int {
	return len(r.Records)
}

func (r *PingRecord) PacketLoss() float64 {
	if r.Count == 0 {
		return 0
	}

	successRate := float64(r.Success) / float64(r.Count)
	return 100.0 * (1.0 - successRate)
}

func (r *PingRecord) TimeCostMs() float64 {
	duration := r.Finish.Sub(r.Start)
	return float64(duration) / float64(time.Millisecond)
}
