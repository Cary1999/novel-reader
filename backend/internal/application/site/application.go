package siteapp

import (
	sitecommand "novel-reader/backend/internal/application/site/command"
	sitequery "novel-reader/backend/internal/application/site/query"
	siterepository "novel-reader/backend/internal/domain/site/repository"
	siteservice "novel-reader/backend/internal/domain/site/service"
)

type Commands struct {
	UpdateSettings *sitecommand.UpdateSettingsHandler
	UpdateIcon     *sitecommand.UpdateIconHandler
}

type Queries struct {
	GetSettings *sitequery.GetSettingsHandler
}

func NewCommands(repo siterepository.SiteRepository, iconStore sitecommand.IconStore) *Commands {
	service := siteservice.NewSiteService()
	return &Commands{
		UpdateSettings: sitecommand.NewUpdateSettingsHandler(repo, service),
		UpdateIcon:     sitecommand.NewUpdateIconHandler(repo, service, iconStore),
	}
}

func NewQueries(repo siterepository.SiteRepository) *Queries {
	return &Queries{
		GetSettings: sitequery.NewGetSettingsHandler(repo, siteservice.NewSiteService()),
	}
}
