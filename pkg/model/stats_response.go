package model

type StatsResponse struct {
	Platform             string         `json:"platform"`
	Username             string         `json:"username"`
	TotalSolved          int            `json:"totalSolved,omitempty"`
	Rating               int            `json:"rating,omitempty"`
	Streak               int            `json:"streak,omitempty"`
	EasySolved           int            `json:"easySolved,omitempty"`
	MediumSolved         int            `json:"mediumSolved,omitempty"`
	HardSolved           int            `json:"hardSolved,omitempty"`
	ContestsParticipated int            `json:"contestsParticipated,omitempty"`
	MaxRating    		 int    		`json:"maxRating,omitempty"`
	QuestionsByType      map[string]int `json:"questionsByType,omitempty"`
	GlobalRank   		 int    		`json:"globalRank,omitempty"`
	CountryRank  		 int   			`json:"countryRank,omitempty"`

	// Add these two fields
	Badges               []string       `json:"badges,omitempty"`
	Certifications       int            `json:"certifications,omitempty"`
	CertificationLinks   []string       `json:"certificationLinks,omitempty"`
	CodingScore          int            `json:"codingScore,omitempty"`
	Error                string         `json:"error,omitempty"`
}
