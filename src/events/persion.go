package events

import (
	"fmt"
)

type Man struct {
}

func (man *Man) Say() string {
	fmt.Println("I am man!")
	return "ok"
}

func (man *Man) Eat(food string) string {
	fmt.Println("I (man)like (", food, ")!")
	return "ok"
}

type WoMan struct {
}

func (woman *WoMan) Say() string {
	fmt.Println("I am woman!")
	return "ok"
}

func (woman *WoMan) Eat(food string) string {
	fmt.Println("I (woman)like (", food, ")!")
	return "ok"
}
