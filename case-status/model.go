package case_status

type CaseStatus struct {
	UserId		string	`json:"userId"`
	CorpId		string	`json:"corpId"`
	CaseId		string	`json:"caseId"`
	Status		int		`json:"status"`
}