package main

import (
	"fmt"
	"math/big"
)

type BigCalculator struct{}

func (bC BigCalculator) Add(a, b *big.Int) *big.Int {
	return new(big.Int).Add(a, b)
}

func (bC BigCalculator) Sub(a, b *big.Int) *big.Int {
	return new(big.Int).Sub(a, b)
}

func (bC BigCalculator) Mul(a, b *big.Int) *big.Int {
	return new(big.Int).Mul(a, b)
}

func (bC BigCalculator) Div(a, b *big.Int) *big.Int {
	return new(big.Int).Div(a, b)
}

func main() {
	bC := BigCalculator{}
	a := new(big.Int).Exp(big.NewInt(2), big.NewInt(30), nil)
	b := new(big.Int).Exp(big.NewInt(2), big.NewInt(25), nil)

	fmt.Println(bC.Add(a, b))
	fmt.Println(bC.Sub(a, b))
	fmt.Println(bC.Mul(a, b))
	fmt.Println(bC.Div(a, b))
}

//done
