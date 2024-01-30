package manager

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
	"ueba-profiling/external"
	"ueba-profiling/profile"
	"ueba-profiling/view"
)

type (
	JobManager struct {
		RunningWorkers map[string]*JobMetadata
		FailedWorkers  map[string]*FailedJobMetadata
		logger         *logrus.Entry
	}
	JobMetadata struct {
		worker    profile.IWorker
		JobConfig *view.JobConfig
	}
	FailedJobMetadata struct {
		JobConfig *view.JobConfig
		err       error
	}
)

func NewJobManager() *JobManager {
	return &JobManager{
		RunningWorkers: make(map[string]*JobMetadata),
		FailedWorkers:  make(map[string]*FailedJobMetadata),
		logger:         logrus.WithField("manager", "job"),
	}
}

func (m *JobManager) Run() {
	go func() {
		m.pullJobs()
		ticker := time.NewTicker(10 * time.Minute)
		for {
			select {
			case <-ticker.C:
				m.pullJobs()
			}
		}
	}()
}

func (m *JobManager) pullJobs() {
	jobHub := external.NewJobHub()
	jobs, err := jobHub.GetJobs()
	if err != nil {
		m.logger.Errorf("error in pulling jobs from JobHub: %v", err)
	}
	for _, job := range jobs {
		if _, ok := m.RunningWorkers[job.ID]; ok {
			continue
		}
		delete(m.FailedWorkers, job.ID)
		err := m.CreateJob(job)
		if err != nil {
			m.logger.Errorf("error in create profiling jobs: %v", err)
		}
	}
}

func (m *JobManager) CreateJob(jobConfig *view.JobConfig) error {
	if _, ok := m.RunningWorkers[jobConfig.ID]; ok {
		return fmt.Errorf("job with id %v already running", jobConfig.ID)
	}

	worker, err := profile.NewProfilingWorker(jobConfig)
	if err != nil {
		m.FailedWorkers[jobConfig.ID] = &FailedJobMetadata{
			JobConfig: jobConfig,
			err:       err,
		}
		return err
	}
	m.logger.Infof("Start Running Profiling Job with ID %v, type %v", jobConfig.ID, jobConfig.ProfileConfig.ProfileType)
	worker.Start()
	m.RunningWorkers[jobConfig.ID] = &JobMetadata{
		worker:    worker,
		JobConfig: jobConfig,
	}

	return nil
}

func (m *JobManager) GetJobInfo() map[string]*view.JobConfig {
	return nil
}
