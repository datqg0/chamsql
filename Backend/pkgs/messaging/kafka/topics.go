package kafka_config

const (
	// Exam domain events
	TopicExamEvents = "chamsql-exam-events-v1"

	// Submission domain events
	TopicSubmissionEvents = "chamsql-submission-events-v1"

	// Submission grading (Phase 4)
	TopicStudentSubmission = "chamsql-student-submission-v1" // Students submit solutions
	TopicSubmissionGraded  = "chamsql-submission-graded-v1"  // Grading results returned
)

const (
	// Consumer groups
	GroupExamWorkers       = "chamsql-exam-workers"
	GroupSubmissionWorkers = "chamsql-submission-workers"
	GroupGradingWorkers    = "chamsql-grading-workers" // Phase 4
)
