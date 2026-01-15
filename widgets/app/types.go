package app

// DSL the app DSL.
//
// This structure is used by `widgets/app` to build Xgen settings (e.g. `yao.app.Xgen`).
// It is parsed from `app.yao` / `app.jsonc` / `app.json` but is NOT the same as `share.AppInfo`
// (which is used by the engine startup).
//
// Keep this struct minimal and frontend-oriented: only include fields needed by Xgen and
// `widgets/app` processes.
type DSL struct {
	Name        string      `json:"name,omitempty"`
	Short       string      `json:"short,omitempty"`
	Version     string      `json:"version,omitempty"`
	Description string      `json:"description,omitempty"`
	Theme       string      `json:"theme,omitempty"`
	Lang        string      `json:"lang,omitempty"`
	Sid         string      `json:"sid,omitempty"`
	Logo        string      `json:"logo,omitempty"`
	Favicon     string      `json:"favicon,omitempty"`
	Menu        MenuDSL     `json:"menu,omitempty"`
	AdminRoot   string      `json:"adminRoot,omitempty"`
	Optional    OptionalDSL `json:"optional,omitempty"`

	// Token is the Xgen-side token persistence config. It is exposed by `yao.app.Xgen`.
	// (The login process still lives in `logins/*.login.*`.)
	Token   OptionalDSL `json:"token,omitempty"`
	Setting string      `json:"setting,omitempty"` // custom setting process
	Setup   string      `json:"setup,omitempty"`   // setup process
}

// MenuDSL the menu DSL
type MenuDSL struct {
	Process string        `json:"process,omitempty"`
	Args    []interface{} `json:"args,omitempty"`
}

// OptionalDSL the Optional DSL
type OptionalDSL map[string]interface{}

// CFUN cloud function
type CFUN struct {
	Method string        `json:"method"`
	Args   []interface{} `json:"args,omitempty"`
}
