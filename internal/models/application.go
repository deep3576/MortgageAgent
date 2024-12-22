package models

import "time"

type Application struct {
	ID              int
	BrokerID        int
	ApplicationType string
	AssignedAdminID *int
	CreatedAt       time.Time
}

type DocumentInfo struct {
	Category string
	FilePath string
}

// ApplicationWithDocuments holds application data along with its associated documents.
type ApplicationWithDocuments struct {
	ID              int
	BrokerID        int
	ApplicationType string
	CreatedAt       time.Time
	Documents       []DocumentInfo
}
