package kafka_config

const (
	// Exam domain events
	TopicExamEvents = "chamsql-exam-events-v1"

	// Submission domain events
	TopicSubmissionEvents = "chamsql-submission-events-v1"
)

const (
	// Consumer groups
	GroupExamWorkers       = "chamsql-exam-workers"
	GroupSubmissionWorkers = "chamsql-submission-workers"
)
