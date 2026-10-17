package parser

// TokenType описывает тип токена
type TokenType int

const (
	TokenArg               TokenType = iota // Обычный аргумент
	TokenRedirectOut                        // >
	TokenRedirectAppend                     // >>
	TokenRedirectIn                         // <
	TokenRedirectErr                        // 2>
	TokenRedirectAppendErr                  // 2>>
	TokenPipe                               // |
)

// Token представляет один токен
type Token struct {
	Type  TokenType
	Value string
}

// Redirection описывает перенаправление потока.
type Redirection struct {
	Type     TokenType // Тип перенаправления.
	FilePath string
}

// Command представляет одну команду в pipeline.
type Command struct {
	Name         string        // Имя команды (например, "echo")
	Args         []string      // Аргументы (например, ["hello"])
	Redirections []Redirection // Перенаправления (если есть)
}

// Pipeline — одна или несколько команд, соединённых `|`.
type Pipeline struct {
	Commands   []*Command
	Background bool // завершающий `&` относится ко всему pipeline
}
