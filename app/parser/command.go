package parser

// TokenType описывает тип токена
type TokenType int

const (
	TokenArg            TokenType = iota // Обычный аргумент
	TokenRedirectOut                     // >
	TokenRedirectAppend                  // >>
	TokenRedirectIn                      // <
)

// Token представляет один токен
type Token struct {
	Type  TokenType
	Value string
}

// Redirection описывает перенаправление потока (>, >>, <)
type Redirection struct {
	Type     TokenType // TokenRedirectOut, TokenRedirectAppend, TokenRedirectIn
	FilePath string
}

// Command представляет собой распарсенную строку ввода.
type Command struct {
	Name         string        // Имя команды (например, "echo")
	Args         []string      // Аргументы (например, ["hello"])
	Redirections []Redirection // Перенаправления (если есть)
}
