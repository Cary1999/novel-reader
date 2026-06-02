package identityapp

import (
	identitycommand "novel-reader/backend/internal/application/identity/command"
	identityquery "novel-reader/backend/internal/application/identity/query"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	identityservice "novel-reader/backend/internal/domain/identity/service"
)

type TokenIssuer = identitycommand.TokenIssuer

type Commands struct {
	Register       *identitycommand.RegisterHandler
	Login          *identitycommand.LoginHandler
	UpdateNickname *identitycommand.UpdateNicknameHandler
	ChangePassword *identitycommand.ChangePasswordHandler
}

type Queries struct {
	CurrentUser *identityquery.CurrentUserHandler
}

func NewCommands(users identityrepository.UserRepository, tokens TokenIssuer) *Commands {
	service := identityservice.NewIdentityService()
	return &Commands{
		Register:       identitycommand.NewRegisterHandler(users, service),
		Login:          identitycommand.NewLoginHandler(users, tokens),
		UpdateNickname: identitycommand.NewUpdateNicknameHandler(users, service),
		ChangePassword: identitycommand.NewChangePasswordHandler(users, service),
	}
}

func NewQueries(users identityrepository.UserRepository) *Queries {
	return &Queries{
		CurrentUser: identityquery.NewCurrentUserHandler(users),
	}
}
