package models

type UserCredentials struct {
	ID        int
	UserID    int
	Login     string
	Password  string
	MetaInfo  string
	UpdatedAt string
}

type TextData struct {
	ID        int
	UserID    int
	Data      string
	MetaInfo  string
	UpdatedAt string
}

type CreditCardData struct {
	ID             int
	UserID         int
	CardNumber     string
	ExpirationDate string
	CVV            int
	MetaInfo       string
	UpdatedAt      string
}

type FilesData struct {
	ID        int
	UserID    int
	Path      string
	MetaInfo  string
	UpdatedAt string
}

type SyncData struct {
	UserCredentials []UserCredentials
	TextData        []TextData
	CreditCardData  []CreditCardData
	FilesData       []FilesData
}
