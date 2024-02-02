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

func (uc *UserCredentials) Fields() ([]string, []interface{}) {
	return []string{"id", "user_id", "login", "password", "meta_info", "updated_at"},
		[]interface{}{uc.ID, uc.UserID, uc.Login, uc.Password, uc.MetaInfo, uc.UpdatedAt}
}

func (td *TextData) Fields() ([]string, []interface{}) {
	return []string{"id", "user_id", "data", "meta_info", "updated_at"},
		[]interface{}{td.ID, td.UserID, td.Data, td.MetaInfo, td.UpdatedAt}
}

func (ccd *CreditCardData) Fields() ([]string, []interface{}) {
	return []string{"id", "user_id", "card_number", "expiration_date", "cvv", "meta_info", "updated_at"},
		[]interface{}{ccd.ID, ccd.UserID, ccd.CardNumber, ccd.ExpirationDate, ccd.CVV, ccd.MetaInfo, ccd.UpdatedAt}
}

func (fd *FilesData) Fields() ([]string, []interface{}) {
	return []string{"id", "user_id", "path", "meta_info", "updated_at"},
		[]interface{}{fd.ID, fd.UserID, fd.Path, fd.MetaInfo, fd.UpdatedAt}
}
