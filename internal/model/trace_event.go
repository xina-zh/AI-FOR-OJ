package model

type TraceEvent struct {
	CreatedModel

	ExperimentRunID uint          `gorm:"column:experiment_run_id;not null;index:idx_trace_run;uniqueIndex:idx_trace_run_sequence" json:"experiment_run_id"`
	SequenceNo      int           `gorm:"column:sequence_no;not null;uniqueIndex:idx_trace_run_sequence" json:"sequence_no"`
	StepType        string        `gorm:"column:step_type;type:varchar(64);not null;index" json:"step_type"`
	Content         string        `gorm:"type:longtext;not null" json:"content"`
	Metadata        string        `gorm:"type:longtext;not null" json:"metadata"`
	ExperimentRun   ExperimentRun `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"experiment_run,omitempty"`
}

func (TraceEvent) TableName() string {
	return "trace_events"
}
