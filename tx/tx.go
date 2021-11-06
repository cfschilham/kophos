package tx

const (
	SigTypeSHA256 = "SHA256"
	SigType
)

type TxHeader struct {
	To     int
	From   int
	Amount int
}

type Sig struct {
	Type string
	Value [32]byte
}

type Tx struct {
	TxHeader
	Sig [32]byte
}

//func New(header TxHeader) Tx {
//	rsa.SignPKCS1v15()
//	crypto.RegisterHash()
//	return Tx{TxHeader: header, Sig}
//}