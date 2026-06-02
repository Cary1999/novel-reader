package entity

type Upload struct {
	ActorUserID      int64
	OriginalFilename string
	StoredPath       string
	FileSize         int64
	Status           string
}
