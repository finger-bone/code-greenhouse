package query

import "judge/schema"

func (this *r) TestingsByRepository(args struct{ RepositoryId string }) ([]schema.Testing, error) {
	var r []schema.Testing
	if err := this.db.Where("repository_id = ?", args.RepositoryId).Find(&r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func (this *r) TestingsByStage(args struct {
	RepositoryId string
	Stage        int32
}) ([]schema.Testing, error) {
	var r []schema.Testing
	if err := this.db.Where("repository_id = ? AND stage = ?", args.RepositoryId, args.Stage).Find(&r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func (this *r) Testing(args struct {
	RepositoryId string
	Serial       int32
}) (*schema.Testing, error) {
	var r schema.Testing
	if err := this.db.Where("repository_id = ? AND serial = ?", args.RepositoryId, args.Serial).First(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}
