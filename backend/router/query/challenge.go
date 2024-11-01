package query

import (
	"judge/challenge"
)

func (this *r) Challenges() ([]challenge.Challenge, error) {
	return challenge.ParseAllChallenges(this.logger, &this.config.Challenge)
}

func (this *r) Challenge(args struct{ FolderName string }) (*challenge.Challenge, error) {
	return challenge.ParseChallenge(this.logger, &this.config.Challenge, args.FolderName)
}
