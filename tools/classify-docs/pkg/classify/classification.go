package classify

type DocumentClassification struct {
	File                    string   `json:"file"`
	Classification          string   `json:"classification"`
	Confidence              string   `json:"confidence"`
	MarkingsFound           []string `json:"markings_found"`
	ClassificationRationale string   `json:"classification_rationale"`
}
