package domain

type Rule struct {
	Roles   string   `json:"roles"`
	Actions []string `json:"actions"`
	Effect  string   `json:"effect,omitempty"`
}

type ResourcePolicy struct {
	Version  string `json:"version,omitempty"`
	Resource string `json:"resource"`
	Scope    string `json:"scope,omitempty"`
	Rules    []Rule `json:"rules"`
}

type Policy struct {
	APIVersion     string         `json:"apiVersion,omitempty"`
	ResourcePolicy ResourcePolicy `json:"resourcePolicy"`
}

type CerbosPolicy struct {
	PolicyKind string   `json:"policyKind,omitempty"`
	Policies   []Policy `json:"policies"`
}

type Condition struct {
	Match Match `json:"match"`
}

type Match struct {
	Expr string `json:"expr"`
}

type Cerbos struct {
	ID string `gorm:"type:char(36);primary_key" json:"id"`
}

type CerbosResponse struct {
	ID string `json:"id"`
}

type CerbosPolies struct {
	Policy string `json:"policy"`
}
