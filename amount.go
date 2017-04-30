package main

type Amount struct {
	amount string
}

func NewAmount(amount string) *Amount {
	return &Amount{amount: amount}
}
