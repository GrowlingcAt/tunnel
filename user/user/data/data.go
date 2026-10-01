package data

import "database/sql"

type IData interface {
	NewUserData() IUserData
	NewGitlabUserData() IGitlabUserData
	NewOfficialUserData() IOfficialUserData
	NewGithubUserData() IGithubUserData
}

type data struct {
	db *sql.DB
}

func NewData(db *sql.DB) IData {
	return &data{
		db: db,
	}
}
func (d *data) NewUserData() IUserData {
	return &userData{
		table: TBL_USER,
		db:    d.db,
	}
}
func (d *data) NewGitlabUserData() IGitlabUserData {
	return &gitlabUserData{
		table: TBL_GITLAB_USER,
		db:    d.db,
	}

}
func (d *data) NewOfficialUserData() IOfficialUserData {
	return &officialUserData{
		table: TBL_OFFICIAL_USER,
		db:    d.db,
	}
}
func (d *data) NewGithubUserData() IGithubUserData {
	return &githubUserData{
		table: TBL_GITHUB_USER,
		db:    d.db,
	}
}
