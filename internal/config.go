package internal

const EnvironmentGenerator = "environment_generator"
const HttpAction = "http_action"

type GeneratorConfig struct {
	Type      string            `json:"type"`
	Arguments map[string]string `json:"arguments"`
}

type MoveConfig struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type ActionConfig struct {
	Type      string                 `json:"type"`
	Arguments map[string]interface{} `json:"Arguments"`
}

type CleanConfig struct {
	Folders []string `json:"folders,omitempty"`
	Files   []string `json:"files,omitempty"`
}

type StepConfig struct {
	Generate []GeneratorConfig `json:"generate,omitempty"`
	Move     []MoveConfig      `json:"move,omitempty"`
	Action   []ActionConfig    `json:"action,omitempty"`
	Clean    CleanConfig       `json:"clean,omitempty"`
}

type FtpSyncConfig struct {
	Source          string   `json:"source"`
	Destination     string   `json:"destination"`
	LogFileDest     string   `json:"log_file_dest"`
	IgnoreList      []string `json:"ignore_list"`
	DefaultFileMode string   `json:"default_file_mode"`
	DefaultDirMode  string   `json:"default_dir_mode"`
	FtpConfig       struct {
		Host     string `json:"host"`
		User     string `json:"user"`
		Password string `json:"password"`
	} `json:"ftp_config"`
}

type FtpConfig struct {
	Before          StepConfig    `json:"before,omitempty"`
	Sync            FtpSyncConfig `json:"sync"`
	Folders         []string      `json:"folders,omitempty"`
	ReadableFolders []string      `json:"readable_folders,omitempty"`
	After           StepConfig    `json:"after,omitempty"`
}
