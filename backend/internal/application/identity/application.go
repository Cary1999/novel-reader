package identityapp

import (
	identitycommand "novel-reader/backend/internal/application/identity/command"
	identityquery "novel-reader/backend/internal/application/identity/query"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	identityservice "novel-reader/backend/internal/domain/identity/service"
)

type UserTokenIssuer = identitycommand.UserTokenIssuer
type OperatorTokenIssuer = identitycommand.OperatorTokenIssuer

type Commands struct {
	Register                *identitycommand.RegisterHandler
	Login                   *identitycommand.LoginHandler
	AdminLogin              *identitycommand.AdminLoginHandler
	UpdateNickname          *identitycommand.UpdateNicknameHandler
	UpdateAvatar            *identitycommand.UpdateAvatarHandler
	ChangePassword          *identitycommand.ChangePasswordHandler
	SubmitAuthorApplication *identitycommand.SubmitAuthorApplicationHandler
	ReviewAuthorApplication *identitycommand.ReviewAuthorApplicationHandler
	PromoteUserToAuthor     *identitycommand.PromoteUserToAuthorHandler
	CreateOperator          *identitycommand.CreateOperatorHandler
	UpdateOperator          *identitycommand.UpdateOperatorHandler
}

type Queries struct {
	CurrentUser            *identityquery.CurrentUserHandler
	GetUser                *identityquery.GetUserHandler
	CurrentOperator        *identityquery.CurrentOperatorHandler
	MyAuthorApplication    *identityquery.MyAuthorApplicationHandler
	ListAuthorApplications *identityquery.ListAuthorApplicationsHandler
	ListUsers              *identityquery.ListUsersHandler
	ListOperators          *identityquery.ListOperatorsHandler
}

func NewCommands(users identityrepository.UserRepository, operators identityrepository.OperatorRepository, applications identityrepository.AuthorApplicationRepository, userTokens UserTokenIssuer, operatorTokens OperatorTokenIssuer, avatarStore identitycommand.AvatarStore) *Commands {
	service := identityservice.NewIdentityService()
	return &Commands{
		Register:                identitycommand.NewRegisterHandler(users, service),
		Login:                   identitycommand.NewLoginHandler(users, userTokens),
		AdminLogin:              identitycommand.NewAdminLoginHandler(operators, operatorTokens),
		UpdateNickname:          identitycommand.NewUpdateNicknameHandler(users, service),
		UpdateAvatar:            identitycommand.NewUpdateAvatarHandler(users, avatarStore),
		ChangePassword:          identitycommand.NewChangePasswordHandler(users, service),
		SubmitAuthorApplication: identitycommand.NewSubmitAuthorApplicationHandler(users, applications, service),
		ReviewAuthorApplication: identitycommand.NewReviewAuthorApplicationHandler(users, applications, service),
		PromoteUserToAuthor:     identitycommand.NewPromoteUserToAuthorHandler(users),
		CreateOperator:          identitycommand.NewCreateOperatorHandler(operators, service),
		UpdateOperator:          identitycommand.NewUpdateOperatorHandler(operators, service),
	}
}

func NewQueries(users identityrepository.UserRepository, operators identityrepository.OperatorRepository, applications identityrepository.AuthorApplicationRepository) *Queries {
	return &Queries{
		CurrentUser:            identityquery.NewCurrentUserHandler(users),
		GetUser:                identityquery.NewGetUserHandler(users),
		CurrentOperator:        identityquery.NewCurrentOperatorHandler(operators),
		MyAuthorApplication:    identityquery.NewMyAuthorApplicationHandler(applications),
		ListAuthorApplications: identityquery.NewListAuthorApplicationsHandler(applications),
		ListUsers:              identityquery.NewListUsersHandler(users),
		ListOperators:          identityquery.NewListOperatorsHandler(operators),
	}
}
