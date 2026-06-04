package bookshelfapp

import (
	bookshelfrepository "novel-reader/backend/internal/domain/bookshelf/repository"
	bookshelfservice "novel-reader/backend/internal/domain/bookshelf/service"
)

type Commands struct {
	CreateGroup  *CreateGroupHandler
	RenameGroup  *RenameGroupHandler
	ReorderGroup *ReorderGroupHandler
	PinGroup     *PinGroupHandler
	DeleteGroup  *DeleteGroupHandler
	AddBook      *AddBookHandler
	UpdateBook   *UpdateBookHandler
	RemoveBook   *RemoveBookHandler
	BatchManage  *BatchManageHandler
}

type Queries struct {
	FindGroup   *FindGroupHandler
	ListGroups  *ListGroupsHandler
	ListEntries *ListEntriesHandler
	FindEntry   *FindEntryHandler
}

func NewCommands(repo bookshelfrepository.Repository) *Commands {
	service := bookshelfservice.NewService()
	return &Commands{
		CreateGroup:  NewCreateGroupHandler(repo, service),
		RenameGroup:  NewRenameGroupHandler(repo, service),
		ReorderGroup: NewReorderGroupHandler(repo, service),
		PinGroup:     NewPinGroupHandler(repo),
		DeleteGroup:  NewDeleteGroupHandler(repo),
		AddBook:      NewAddBookHandler(repo),
		UpdateBook:   NewUpdateBookHandler(repo),
		RemoveBook:   NewRemoveBookHandler(repo),
		BatchManage:  NewBatchManageHandler(repo, service),
	}
}

func NewQueries(repo bookshelfrepository.Repository) *Queries {
	service := bookshelfservice.NewService()
	return &Queries{
		FindGroup:   NewFindGroupHandler(repo),
		ListGroups:  NewListGroupsHandler(repo),
		ListEntries: NewListEntriesHandler(repo, service),
		FindEntry:   NewFindEntryHandler(repo),
	}
}
