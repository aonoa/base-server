package service

import (
	"fmt"
	"time"
)

var DefaultJobs map[string]JobFunc

type JobFunc func()

type JobService struct {
	//uc *biz.GreeterUsecase
}

func NewJobService() *JobService {
	job := &JobService{
		//uc: uc,
	}
	return job
}

func (s *JobService) Init() {
	DefaultJobs = map[string]JobFunc{
		"one": s.DoMyWork,
		"two": s.DoOtherWork,
	}
}

func (s *JobService) DoMyWork() {
	fmt.Printf("当前时间 %v \n", time.Now().Unix())
}

func (s *JobService) DoOtherWork() {
	fmt.Printf("当前时间2 %v \n", time.Now().Unix())
}
