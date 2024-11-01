package query

import "judge/schema"

type UserResponse struct {
	Subject    string
	Provider   string
	CreateTime string
	UpdateTime string
	Attributes []schema.UserAttribute
}

func (this *r) User(args struct {
	Subject  string
	Provider string
}) (*UserResponse, error) {
	db := this.db
	user := new(schema.User)
	db.Where("subject = ? AND provider = ?", args.Subject, args.Provider).First(&schema.User{})
	if db.Error != nil {
		return nil, db.Error
	}
	attributes := make([]schema.UserAttribute, 0)
	db.Where("subject = ? AND provider = ?", args.Subject, args.Provider).Find(&attributes)
	if db.Error != nil {
		return nil, db.Error
	}
	return &UserResponse{
		Subject:    user.Subject,
		Provider:   user.Provider,
		CreateTime: user.CreateTime,
		UpdateTime: user.UpdateTime,
		Attributes: attributes,
	}, nil
}
