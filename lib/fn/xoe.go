package fn

import "log"

func Poe(err error) bool {
	if err != nil {
		log.Println("Poe:", err)
		panic(err)
	}
	return false
}

func Poe1[T1 any](v1 T1, err error) T1 {
	Poe(err)
	return v1
}

func Loe(err error) bool {
	if err != nil {
		log.Println("Loe:", err)
	}
	return true
}

func Loe1[T1 any](v1 T1, err error) T1 {
	Loe(err)
	return v1
}
