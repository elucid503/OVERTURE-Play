package main

import (
	"OVERTURE/Play/Public"
	"fmt"
)

func main() {

	// Simple Tester

	Res, Err := Public.Info("https://youtube.com/watch?v=dQw4w9WgXcQ", &Public.InfoOptions{ GetHLSFormats: true }, nil, nil);

	if Err != nil {

		fmt.Println("Info Fetch Error:", Err);
		return;

	}

	fmt.Println(Res.HLSFormats);

}