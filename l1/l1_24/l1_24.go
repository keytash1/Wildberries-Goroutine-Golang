package main

import (
	"fmt"
	"math"
)

type Point struct {
	x float64
	y float64
}

func NewPoint(x, y float64) Point {
	point := Point{
		x: x,
		y: y,
	}
	return point
}

func (point Point) Distance(other Point) float64 {
	return math.Sqrt(math.Pow(point.x-other.x, 2) + math.Pow(point.y-other.y, 2))
}

func main() {
	firstpoint := NewPoint(0, 0)
	secondpoint := NewPoint(1, 1)
	fmt.Println(firstpoint.Distance(secondpoint))
}

//done
