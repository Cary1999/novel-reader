package model

import uploadentity "novel-reader/backend/internal/domain/upload/entity"

type Upload struct {
	ID               int64
	ActorUserID      int64
	OriginalFilename string
	StoredPath       string
	FileSize         int64
	Status           string
	ErrorMessage     string
}

func (m *Upload) FromEntity(e *uploadentity.Upload) {
	m.ActorUserID = e.ActorUserID
	m.OriginalFilename = e.OriginalFilename
	m.StoredPath = e.StoredPath
	m.FileSize = e.FileSize
	m.Status = e.Status
}

func (m Upload) ToEntity() uploadentity.Upload {
	return uploadentity.Upload{
		ActorUserID:      m.ActorUserID,
		OriginalFilename: m.OriginalFilename,
		StoredPath:       m.StoredPath,
		FileSize:         m.FileSize,
		Status:           m.Status,
	}
}
