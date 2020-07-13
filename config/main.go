package main

type Queue struct {
	Name            string
	MaxLimit        int
	MonitorInterval int
}

type QueueOption func(*Queue)

func WithMaxLimit(max int) QueueOption {
	return func(q *Queue) {
		q.MaxLimit = max
	}
}

func WithMonitorInterval(seconds int) QueueOption {
	return func(q *Queue) {
		q.MonitorInterval = seconds
	}
}

func NewQueue(name string, options ...QueueOption) *Queue {
	// 默认参数
	queue := &Queue{name, 10, 5}

	for _, o := range options {
		o(queue)
	}
	return queue
}

func main() {
	NewQueue("abc", WithMaxLimit(10), WithMonitorInterval(2))
}
