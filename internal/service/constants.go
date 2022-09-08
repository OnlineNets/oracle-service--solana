package service

type TokenType uint32

const (
	TypeNative TokenType = iota
	TypeFT
	TypeNFT
)
