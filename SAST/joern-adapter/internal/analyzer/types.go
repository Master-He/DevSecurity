package analyzer

import "time"

const EngineVersion = "4.0.674"

type Finding struct {
	RuleID         string `json:"ruleId"`
	Title          string `json:"title"`
	Severity       string `json:"severity"`
	Confidence     string `json:"confidence"`
	Category       string `json:"category"`
	CWE            string `json:"cwe"`
	Message        string `json:"message"`
	FilePath       string `json:"filePath"`
	StartLine      int    `json:"startLine"`
	EndLine        int    `json:"endLine"`
	CodeSnippet    string `json:"codeSnippet"`
	Recommendation string `json:"recommendation"`
}

type Result struct {
	Engine        string    `json:"engine"`
	EngineVersion string    `json:"engineVersion"`
	SourceMode    string    `json:"sourceMode"`
	Findings      []Finding `json:"findings"`
	Diagnostics   []string  `json:"diagnostics"`
	CompletedAt   time.Time `json:"completedAt"`
}

type ruleMetadata struct {
	title          string
	severity       string
	confidence     string
	category       string
	cwe            string
	message        string
	recommendation string
}

var rules = map[string]ruleMetadata{
	"SAST-JAVA-001": {
		title:          "潜在命令执行",
		severity:       "HIGH",
		confidence:     "MEDIUM",
		category:       "command-injection",
		cwe:            "CWE-78",
		message:        "代码调用了 Runtime.exec 或 ProcessBuilder.start，请确认命令参数不来自不可信输入。",
		recommendation: "避免将外部输入拼接为 shell 命令；使用参数化进程 API、固定命令白名单和严格输入校验。",
	},
	"SAST-JAVA-002": {
		title:          "使用弱密码学算法",
		severity:       "MEDIUM",
		confidence:     "MEDIUM",
		category:       "weak-cryptography",
		cwe:            "CWE-327",
		message:        "代码使用 MD5、SHA-1、DES 或 RC4 等不推荐算法。",
		recommendation: "密码摘要使用 SHA-256 或更强算法；口令存储使用 Argon2id、scrypt 或 PBKDF2；对称加密使用 AES-GCM。",
	},
	"SAST-JAVA-003": {
		title:          "不安全 Java 反序列化",
		severity:       "HIGH",
		confidence:     "MEDIUM",
		category:       "unsafe-deserialization",
		cwe:            "CWE-502",
		message:        "代码调用 ObjectInputStream.readObject，反序列化不可信数据可能导致代码执行。",
		recommendation: "避免反序列化不可信数据；使用允许列表、类型校验或改用安全的 JSON 数据契约。",
	},
}
