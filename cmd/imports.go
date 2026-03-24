package cmd

// Dedektörleri derleme zamanında kaydet (init() ile).
import (
	_ "github.com/cemililik/leakwatch/internal/detector/aws"
	_ "github.com/cemililik/leakwatch/internal/detector/generic"
	_ "github.com/cemililik/leakwatch/internal/detector/privatekey"
)
