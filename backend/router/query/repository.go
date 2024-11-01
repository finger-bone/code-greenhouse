package query

import "judge/schema"

func (this *r) Repositories(args struct {
	Subject  string
	Provider string
}) ([]schema.Repository, error) {
	responses := make([]schema.Repository, 0)
	this.db.Where("subject = ? AND provider = ?", args.Subject, args.Provider).Find(&responses)
	return responses, nil
}

func (this *r) Repository(args struct{ RepositoryId string }) (*schema.Repository, error) {
	response := new(schema.Repository)
	this.db.Where("repository_id = ?", args.RepositoryId).First(response)
	return response, nil
}
