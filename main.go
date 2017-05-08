package main

import (
	"fact/fact"
	"fmt"
)

func main() {
	// We store fact in graph database.
	// Here we use Cayley to store them.
	f := fact.NewFact("./db/ari.db")

	defer f.Close()

	// Declare truth
	f.Let("แมว").Has("ชื่อ", "เหมียว")
	f.Let("หมา").Has("ชื่อ", "โฮ่ง")
	f.Let("เหมียว").Has("สี", "ดำ")

	fmt.Println(f.What("แมว", "ชื่อ"))
	fmt.Println(f.WhoHas("ชื่อ", "เหมียว"))
	fmt.Println(f.WhoHas("ชื่อ", "โฮ่ง"))
	fmt.Println(f.What(f.What("แมว", "ชื่อ"), "สี"))
}
