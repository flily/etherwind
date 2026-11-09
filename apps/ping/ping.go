package ping

import (
	"math"
	"net"
	"time"
)

const (
	DefaultPingRecordCapacity = 3600
)

type PingRecord struct {
	DisplayName string
	Host        string
	IP          net.IP
	Start       time.Time
	Finish      time.Time
	Count       int
	Success     int
	LastTime    float64
	LastTTL     int
	Records     []float64
}

func newPingRecord(cap int) *PingRecord {
	t := time.Now()
	r := &PingRecord{
		Start:    t,
		Finish:   t,
		Count:    0,
		Success:  0,
		LastTime: 0.0,
		LastTTL:  -1,
		Records:  make([]float64, 0, cap),
	}

	return r
}

func (r *PingRecord) WithIP(ip net.IP) *PingRecord {
	r.DisplayName = ip.String()
	r.IP = ip
	return r
}

func NewPingRecordWithIP(ip net.IP, cap int) *PingRecord {
	return newPingRecord(cap).WithIP(ip)
}

func (r *PingRecord) Add(value float64, ttl int) {
	r.Records = append(r.Records, value)
	r.Success++
	r.Count++
	r.LastTime = value
	r.LastTTL = ttl
	r.Finish = time.Now()
}

func (r *PingRecord) AddFailure() {
	r.Count++
	r.LastTime = 0.0
	r.LastTTL = -1
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
