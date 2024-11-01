package schema

type RepositoryTestingSerial struct {
	RepositoryId string `gorm:"primaryKey"`
	NextSerial   int
}

type Testing struct {
	RepositoryId string `gorm:"primaryKey"`
	Serial       int32  `gorm:"primaryKey"`
	Stage        int32
	Status       string
	Message      string
	Log          string
	CreateTime   string
	RunStartTime string
	RunEndTime   string
}

type User struct {
	Subject    string `gorm:"primaryKey"`
	Provider   string `gorm:"primaryKey"`
	CreateTime string
	UpdateTime string
}

type UserAttribute struct {
	Subject  string `gorm:"primaryKey"`
	Provider string `gorm:"primaryKey"`
	Key      string `gorm:"primaryKey"`
	Value    string
}

type UserBasicAuthentication struct {
	Subject            string `gorm:"primaryKey"`
	Provider           string `gorm:"primaryKey"`
	AuthenticationText string
}

type Repository struct {
	RepositoryId        string `gorm:"primaryKey"`
	Subject             string
	Provider            string
	ChallengeFolderName string
	Startpoint          string
	Stage               int32
	TotalStages         int32
	CreateTime          string
	UpdateTime          string
}
